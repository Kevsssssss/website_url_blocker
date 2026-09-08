package service

import (
	"strings"
	"sync"
	"time"

	"github.com/Kevsssssss/website_url_blocker/config"
	"github.com/miekg/dns"
)

// GlobalProxy is the running DNS proxy singleton set by the Windows service.
// It is nil when running in CLI-only mode (service not started).
var GlobalProxy *DNSProxy

// DNSProxy is the local DNS interception server that enforces mutual-block
// group rules automatically by inspecting every DNS query.
type DNSProxy struct {
	mu        sync.Mutex
	sessions  map[string]*SessionEntry // groupName → live session
	quit      chan struct{}
	udpServer *dns.Server
	tcpServer *dns.Server
}

// NewDNSProxy creates a new proxy instance ready to be started.
func NewDNSProxy() *DNSProxy {
	return &DNSProxy{
		sessions: make(map[string]*SessionEntry),
		quit:     make(chan struct{}),
	}
}

// Start binds to 127.0.0.1:53 (UDP + TCP) and launches the idle-expiry loop.
// Returns an error if the port is already in use.
func (p *DNSProxy) Start() error {
	mux := dns.NewServeMux()
	mux.HandleFunc(".", p.handleDNS)

	p.udpServer = &dns.Server{Addr: config.DNSListenAddr, Net: "udp", Handler: mux}
	p.tcpServer = &dns.Server{Addr: config.DNSListenAddr, Net: "tcp", Handler: mux}

	errCh := make(chan error, 2)
	go func() { errCh <- p.udpServer.ListenAndServe() }()
	go func() { errCh <- p.tcpServer.ListenAndServe() }()

	// Give the servers 250ms to fail on bind errors (e.g. port already in use).
	select {
	case err := <-errCh:
		return err
	case <-time.After(250 * time.Millisecond):
		// No immediate error — servers started successfully.
	}

	go p.runIdleExpiry()
	return nil
}

// Stop shuts down the DNS proxy servers and the idle-expiry goroutine.
func (p *DNSProxy) Stop() {
	close(p.quit)
	if p.udpServer != nil {
		p.udpServer.Shutdown()
	}
	if p.tcpServer != nil {
		p.tcpServer.Shutdown()
	}
}

// handleDNS is called for every incoming DNS query.
func (p *DNSProxy) handleDNS(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) == 0 {
		dns.HandleFailed(w, r)
		return
	}

	// Normalise the queried name (strip trailing dot, lowercase, strip www.)
	qname := strings.TrimSuffix(strings.ToLower(r.Question[0].Name), ".")
	domain := normalizeDomain(qname)

	// Load group config from disk on every query (file is tiny, reads are cheap).
	groupsPath, _ := config.GroupsPath()
	groups, _ := ReadGroups(groupsPath)

	// Find whether this domain belongs to a group.
	groupIdx := p.findGroupIndex(domain, groups)
	if groupIdx < 0 {
		// Not a grouped domain — forward transparently.
		p.forward(w, r)
		return
	}

	group := groups[groupIdx]

	// ── Parent override mode ────────────────────────────────────────────────
	// If group-allow set an explicit Active domain, honour it.
	if group.Active != "" {
		p.mu.Lock()
		// Sync in-memory session to match the parent override.
		p.sessions[group.Name] = &SessionEntry{
			ActiveDomain: group.Active,
			LastSeen:     time.Now(),
			Source:       "override",
		}
		p.mu.Unlock()
		go p.persistState()

		if group.Active == domain {
			p.forward(w, r)
		} else {
			p.sendNXDOMAIN(w, r)
		}
		return
	}

	// ── Auto mode ───────────────────────────────────────────────────────────
	// Determine per-group timeout (fall back to global default).
	timeout := time.Duration(config.GroupSessionTimeout) * time.Second
	if group.TimeoutSeconds > 0 {
		timeout = time.Duration(group.TimeoutSeconds) * time.Second
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	state, exists := p.sessions[group.Name]
	now := time.Now()

	// Expire a stale lock.
	if exists && state.ActiveDomain != "" && now.Sub(state.LastSeen) > timeout {
		state.ActiveDomain = ""
		state.Source = ""
	}

	if !exists || state.ActiveDomain == "" {
		// No active site — this domain wins the lock.
		p.sessions[group.Name] = &SessionEntry{
			ActiveDomain: domain,
			LastSeen:     now,
			Source:       "auto",
		}
		go p.persistState()
		p.forward(w, r)
		return
	}

	if state.ActiveDomain == domain {
		// Same domain — refresh the last-seen timestamp.
		state.LastSeen = now
		go p.persistState()
		p.forward(w, r)
		return
	}

	// A different domain in the same group is active — block this request.
	p.sendNXDOMAIN(w, r)
}

// runIdleExpiry periodically scans sessions and releases expired locks.
func (p *DNSProxy) runIdleExpiry() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			groupsPath, _ := config.GroupsPath()
			groups, _ := ReadGroups(groupsPath)

			p.mu.Lock()
			changed := false
			now := time.Now()
			for _, g := range groups {
				state, ok := p.sessions[g.Name]
				if !ok || state.ActiveDomain == "" {
					continue
				}
				// Skip if parent override is active (don't expire those).
				if g.Active != "" {
					continue
				}
				timeout := time.Duration(config.GroupSessionTimeout) * time.Second
				if g.TimeoutSeconds > 0 {
					timeout = time.Duration(g.TimeoutSeconds) * time.Second
				}
				if now.Sub(state.LastSeen) > timeout {
					state.ActiveDomain = ""
					state.Source = ""
					changed = true
				}
			}
			p.mu.Unlock()

			if changed {
				p.persistState()
			}

		case <-p.quit:
			return
		}
	}
}

// persistState writes the current in-memory session map to groups_state.json
// so the CLI (group-list) can display live status without IPC.
func (p *DNSProxy) persistState() {
	p.mu.Lock()
	// Take a snapshot while holding the lock.
	snap := make(map[string]*SessionEntry, len(p.sessions))
	for k, v := range p.sessions {
		entry := *v
		snap[k] = &entry
	}
	p.mu.Unlock()

	path, err := config.GroupsStatePath()
	if err != nil {
		return
	}
	WriteGroupsState(path, snap)
}

// findGroupIndex returns the index in groups where domain is a member, or -1.
func (p *DNSProxy) findGroupIndex(domain string, groups []Group) int {
	for i, g := range groups {
		for _, d := range g.Domains {
			if d == domain {
				return i
			}
		}
	}
	return -1
}

// forward proxies the DNS query to the upstream resolver and returns the result.
func (p *DNSProxy) forward(w dns.ResponseWriter, r *dns.Msg) {
	c := &dns.Client{Timeout: 5 * time.Second}
	resp, _, err := c.Exchange(r, config.DNSUpstream)
	if err != nil || resp == nil {
		dns.HandleFailed(w, r)
		return
	}
	w.WriteMsg(resp)
}

// sendNXDOMAIN replies with a "domain not found" response, effectively blocking
// the queried domain. The browser shows "This site can't be reached".
func (p *DNSProxy) sendNXDOMAIN(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.SetRcode(r, dns.RcodeNameError) // NXDOMAIN
	w.WriteMsg(m)
}
