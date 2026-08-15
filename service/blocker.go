package service

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Kevsssssss/website_url_blocker/config"
)

const (
	sentinelStart = "# --- urlblocker START ---"
	sentinelEnd   = "# --- urlblocker END ---"
)

// ReadBlocklist reads the blocklist.txt file and returns a slice of domains.
// Lines starting with # and blank lines are ignored.
func ReadBlocklist(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("blocklist file not found at: %s", path)
		}
		return nil, fmt.Errorf("could not open blocklist: %w", err)
	}
	defer f.Close()

	var domains []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip any protocol prefix the user may have accidentally included
		line = strings.TrimPrefix(line, "https://")
		line = strings.TrimPrefix(line, "http://")
		line = strings.TrimPrefix(line, "www.")
		line = strings.ToLower(line)
		domains = append(domains, line)
	}
	return domains, scanner.Err()
}

// AddDomainToBlocklist appends a domain to blocklist.txt if not already present.
func AddDomainToBlocklist(path, domain string) error {
	domains, err := ReadBlocklist(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	domain = normalizeDomain(domain)
	for _, d := range domains {
		if d == domain {
			return fmt.Errorf("domain '%s' is already in the blocklist", domain)
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open blocklist for writing: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s\n", domain)
	return err
}

// RemoveDomainFromBlocklist removes a domain from blocklist.txt.
func RemoveDomainFromBlocklist(path, domain string) error {
	domain = normalizeDomain(domain)
	domains, err := ReadBlocklist(path)
	if err != nil {
		return err
	}

	found := false
	var newDomains []string
	for _, d := range domains {
		if d == domain {
			found = true
			continue
		}
		newDomains = append(newDomains, d)
	}
	if !found {
		return fmt.Errorf("domain '%s' not found in blocklist", domain)
	}

	// Rewrite the file with a header comment
	f, err := os.OpenFile(path, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not rewrite blocklist: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprintln(w, "# Website URL Blocker - Parental Control Blocklist")
	fmt.Fprintln(w, "# One domain per line. Lines starting with # are comments.")
	fmt.Fprintln(w)
	for _, d := range newDomains {
		fmt.Fprintln(w, d)
	}
	return w.Flush()
}

// ApplyBlocklist injects all domains from blocklist.txt into the hosts file.
func ApplyBlocklist() error {
	blocklistPath, err := config.BlocklistPath()
	if err != nil {
		return err
	}
	domains, err := ReadBlocklist(blocklistPath)
	if err != nil {
		return err
	}
	return injectHosts(domains)
}

// RemoveBlocklist strips the managed block from the hosts file.
func RemoveBlocklist() error {
	return stripHosts()
}

// injectHosts reads the current hosts file, strips any existing managed block,
// then appends a fresh managed block with the given domains.
func injectHosts(domains []string) error {
	content, err := readHostsFile()
	if err != nil {
		return err
	}

	// Strip existing managed block
	cleaned := stripManagedBlock(content)

	if len(domains) == 0 {
		return writeHostsFile(cleaned)
	}

	// Build new managed block
	var sb strings.Builder
	sb.WriteString(cleaned)
	if !strings.HasSuffix(cleaned, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	sb.WriteString(sentinelStart + "\n")
	for _, domain := range domains {
		d := normalizeDomain(domain)
		sb.WriteString(fmt.Sprintf("127.0.0.1 %s\n", d))
		sb.WriteString(fmt.Sprintf("127.0.0.1 www.%s\n", d))
	}
	sb.WriteString(sentinelEnd + "\n")

	return writeHostsFile(sb.String())
}

// stripHosts removes the managed block from the hosts file.
func stripHosts() error {
	content, err := readHostsFile()
	if err != nil {
		return err
	}
	cleaned := stripManagedBlock(content)
	return writeHostsFile(cleaned)
}

// stripManagedBlock removes lines between (and including) the sentinel markers.
func stripManagedBlock(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inside := false
	for _, line := range lines {
		if strings.TrimSpace(line) == sentinelStart {
			inside = true
			continue
		}
		if strings.TrimSpace(line) == sentinelEnd {
			inside = false
			continue
		}
		if !inside {
			result = append(result, line)
		}
	}
	// Remove trailing blank lines caused by sentinel removal
	for len(result) > 0 && strings.TrimSpace(result[len(result)-1]) == "" {
		result = result[:len(result)-1]
	}
	return strings.Join(result, "\n") + "\n"
}

func readHostsFile() (string, error) {
	data, err := os.ReadFile(config.HostsFilePath)
	if err != nil {
		return "", fmt.Errorf("could not read hosts file (run as Administrator): %w", err)
	}
	return string(data), nil
}

func writeHostsFile(content string) error {
	err := os.WriteFile(config.HostsFilePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("could not write hosts file (run as Administrator): %w", err)
	}
	return nil
}

// IsBlocking returns true if the managed block exists in the hosts file.
func IsBlocking() (bool, error) {
	content, err := readHostsFile()
	if err != nil {
		return false, err
	}
	return strings.Contains(content, sentinelStart), nil
}

// GetBlockedDomainsFromHosts returns the list of domains currently in the hosts file block.
func GetBlockedDomainsFromHosts() ([]string, error) {
	content, err := readHostsFile()
	if err != nil {
		return nil, err
	}
	var domains []string
	lines := strings.Split(content, "\n")
	inside := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == sentinelStart {
			inside = true
			continue
		}
		if trimmed == sentinelEnd {
			break
		}
		if inside && strings.HasPrefix(trimmed, "127.0.0.1 ") {
			parts := strings.Fields(trimmed)
			if len(parts) == 2 && !strings.HasPrefix(parts[1], "www.") {
				domains = append(domains, parts[1])
			}
		}
	}
	return domains, nil
}

func normalizeDomain(domain string) string {
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "www.")
	return strings.ToLower(strings.TrimSpace(domain))
}
