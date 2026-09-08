package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Kevsssssss/website_url_blocker/config"
	blockerservice "github.com/Kevsssssss/website_url_blocker/service"
	"github.com/kardianos/service"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

// Run is the CLI entrypoint. It dispatches subcommands.
func Run(args []string) {
	if len(args) < 1 {
		printHelp()
		return
	}

	cmd := args[0]

	switch cmd {
	case "install":
		runServiceAction("install")
	case "uninstall":
		runServiceAction("uninstall")
	case "start":
		runServiceAction("start")
	case "stop":
		runServiceAction("stop")
	case "status":
		cmdStatus()
	case "add":
		if len(args) < 2 {
			fatalf("Usage: urlblocker add <domain>")
		}
		requirePassword()
		cmdAdd(args[1])
	case "remove":
		if len(args) < 2 {
			fatalf("Usage: urlblocker remove <domain>")
		}
		requirePassword()
		cmdRemove(args[1])
	case "list":
		cmdList()
	case "enable":
		requirePassword()
		cmdEnable()
	case "disable":
		cmdDisable()
	case "setpassword", "changepassword":
		cmdSetPassword()
	case "flush":
		cmdFlush()
	case "help", "--help", "-h":
		printHelp()
	// --- Mutual-block group commands ---
	case "group-add":
		if len(args) < 3 {
			fatalf("Usage: urlblocker group-add <name> <domain1> [domain2 ...]")
		}
		requirePassword()
		cmdGroupAdd(args[1], args[2:])
	case "group-remove":
		if len(args) < 2 {
			fatalf("Usage: urlblocker group-remove <name>")
		}
		requirePassword()
		cmdGroupRemove(args[1])
	case "group-allow":
		if len(args) < 3 {
			fatalf("Usage: urlblocker group-allow <name> <domain>")
		}
		requirePassword()
		cmdGroupAllow(args[1], args[2])
	case "group-reset":
		if len(args) < 2 {
			fatalf("Usage: urlblocker group-reset <name>")
		}
		requirePassword()
		cmdGroupReset(args[1])
	case "group-list":
		cmdGroupList()
	case "group-timeout":
		if len(args) < 3 {
			fatalf("Usage: urlblocker group-timeout <name> <minutes>")
		}
		requirePassword()
		cmdGroupTimeout(args[1], args[2])
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

// --- Command Implementations ---

func cmdAdd(domain string) {
	path, err := config.BlocklistPath()
	must(err)
	must(blockerservice.AddDomainToBlocklist(path, domain))
	fmt.Printf("✓ Added '%s' to blocklist.\n", domain)
	fmt.Println("  Run 'enable' to apply changes now, or wait for the service to reload.")
}

func cmdRemove(domain string) {
	path, err := config.BlocklistPath()
	must(err)
	must(blockerservice.RemoveDomainFromBlocklist(path, domain))
	fmt.Printf("✓ Removed '%s' from blocklist.\n", domain)
	fmt.Println("  Run 'enable' to apply changes now.")
}

func cmdList() {
	path, err := config.BlocklistPath()
	must(err)
	domains, err := blockerservice.ReadBlocklist(path)
	must(err)
	if len(domains) == 0 {
		fmt.Println("Blocklist is empty. Add domains with: add <domain>")
		return
	}
	fmt.Printf("Blocked domains (%d):\n", len(domains))
	for _, d := range domains {
		fmt.Printf("  • %s\n", d)
	}
}

func cmdEnable() {
	requireAdmin()
	fmt.Println("Applying blocklist to hosts file...")
	must(blockerservice.ApplyBlocklist())
	fmt.Println("✓ Blocking enabled. Blocked sites are now unreachable.")
}

func cmdDisable() {
	requireAdmin()
	fmt.Println("Removing blocklist from hosts file...")
	must(blockerservice.RemoveBlocklist())
	fmt.Println("✓ Blocking disabled. Hosts file restored.")
}

func cmdFlush() {
	fmt.Println("Flushing DNS cache...")
	cmd := exec.Command("ipconfig", "/flushdns")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error flushing DNS cache: %v\n", err)
	} else {
		fmt.Println("✓ DNS cache flushed successfully.")
	}
	fmt.Println("\nReminder: If a site is still not blocked, please ensure that")
	fmt.Println("'Secure DNS' (or DNS-over-HTTPS) is turned OFF in your browser settings,")
	fmt.Println("as it bypasses the Windows hosts file.")
}

// --- Mutual-block group command implementations ---

func cmdGroupAdd(name string, domains []string) {
	groupsPath, err := config.GroupsPath()
	must(err)
	blocklistPath, err := config.BlocklistPath()
	must(err)

	must(blockerservice.AddGroup(groupsPath, name, domains))
	fmt.Printf("✓ Group '%s' created with %d domain(s).\n", name, len(domains))

	// Also add every domain to blocklist.txt so they are blocked via hosts file
	// when the service is not running (fallback protection).
	added := 0
	for _, d := range domains {
		if err := blockerservice.AddDomainToBlocklist(blocklistPath, d); err == nil {
			added++
		}
	}
	if added > 0 {
		fmt.Printf("  Added %d domain(s) to blocklist.txt.\n", added)
	}
	fmt.Println("  All domains are blocked by default.")
	fmt.Println("  When the service is running, the first one a user opens will be")
	fmt.Println("  allowed automatically; the rest will be blocked until it goes idle.")
}

func cmdGroupRemove(name string) {
	groupsPath, err := config.GroupsPath()
	must(err)
	must(blockerservice.RemoveGroup(groupsPath, name))
	fmt.Printf("✓ Group '%s' removed.\n", name)
	fmt.Println("  Note: domains remain in blocklist.txt. Remove them manually if desired.")
}

// cmdGroupAllow sets a parent override: forces a specific domain to be the
// active (allowed) site in the group, bypassing the automatic DNS detection.
func cmdGroupAllow(name, domain string) {
	requireAdmin()
	groupsPath, err := config.GroupsPath()
	must(err)
	must(blockerservice.GroupSetActive(groupsPath, name, domain))
	fmt.Printf("✓ Parent override set: '%s' is now the allowed site in group '%s'.\n", domain, name)
	fmt.Println("  All other group members are blocked.")
	fmt.Println("  The DNS proxy will enforce this immediately on the next DNS query.")
	fmt.Println("  Run 'group-reset' to return to automatic detection.")
	must(blockerservice.ApplyBlocklist())
	fmt.Println("✓ Hosts file updated.")
}

// cmdGroupReset clears the parent override and returns the group to automatic
// DNS-detection mode (first site opened wins the lock).
func cmdGroupReset(name string) {
	requireAdmin()
	groupsPath, err := config.GroupsPath()
	must(err)
	must(blockerservice.GroupSetActive(groupsPath, name, ""))
	fmt.Printf("✓ Group '%s' returned to automatic mode — all members are blocked.\n", name)
	fmt.Println("  The first site a user opens will be allowed; the rest stay blocked.")
	must(blockerservice.ApplyBlocklist())
	fmt.Println("✓ Hosts file updated.")
}

// cmdGroupList shows all groups with their live DNS-proxy session state.
func cmdGroupList() {
	groupsPath, err := config.GroupsPath()
	must(err)
	groups, err := blockerservice.ReadGroups(groupsPath)
	must(err)

	if len(groups) == 0 {
		fmt.Println("No mutual-block groups defined. Create one with: group-add <name> <domain1> <domain2> ...")
		return
	}

	// Read live session state written by the DNS proxy.
	statePath, _ := config.GroupsStatePath()
	liveState, _ := blockerservice.ReadGroupsState(statePath)

	fmt.Printf("Mutual-block groups (%d):\n\n", len(groups))
	for _, g := range groups {
		// Determine what is currently active and how.
		activeDomain := ""
		activeLabel := "(none — all blocked)"

		if g.Active != "" {
			// Parent override is set in groups.json.
			activeDomain = g.Active
			activeLabel = g.Active + "  [parent override]"
		} else if liveState != nil {
			if sess, ok := liveState[g.Name]; ok && sess.ActiveDomain != "" {
				activeDomain = sess.ActiveDomain
				ago := time.Since(sess.LastSeen).Truncate(time.Second)
				activeLabel = fmt.Sprintf("%s  [auto, last seen %s ago]", sess.ActiveDomain, ago)
			}
		}

		// Timeout display
		timeoutSec := config.GroupSessionTimeout
		if g.TimeoutSeconds > 0 {
			timeoutSec = g.TimeoutSeconds
		}

		fmt.Printf("  Group  : %s  (timeout: %d sec)\n", g.Name, timeoutSec)
		fmt.Printf("  Active : %s\n", activeLabel)
		fmt.Printf("  Members:\n")
		for _, d := range g.Domains {
			marker := "    •"
			if d == activeDomain {
				marker = "    ▶"
			}
			fmt.Printf("%s %s\n", marker, d)
		}
		fmt.Println()
	}
}

// cmdGroupTimeout sets the per-group inactivity timeout in seconds.
func cmdGroupTimeout(name, secondsStr string) {
	seconds, err := strconv.Atoi(secondsStr)
	if err != nil || seconds < 1 {
		fatalf("Invalid timeout '%s': must be a positive whole number of seconds.", secondsStr)
	}
	groupsPath, err := config.GroupsPath()
	must(err)
	must(blockerservice.SetGroupTimeout(groupsPath, name, seconds))
	fmt.Printf("✓ Group '%s' session timeout set to %d second(s).\n", name, seconds)
	fmt.Println("  The DNS proxy will use this timeout for automatic session expiry.")
}

func cmdStatus() {
	// Service status
	svc, err := blockerservice.NewService()
	if err != nil {
		fmt.Println("Service: (could not connect)")
	} else {
		status, err := svc.Status()
		if err != nil {
			fmt.Println("Service: (unknown)")
		} else {
			switch status {
			case service.StatusRunning:
				fmt.Println("Service: ● RUNNING")
			case service.StatusStopped:
				fmt.Println("Service: ○ STOPPED")
			default:
				fmt.Println("Service: ? UNKNOWN")
			}
		}
	}

	// Hosts file status
	blocking, err := blockerservice.IsBlocking()
	if err != nil {
		fmt.Println("Hosts file: (could not read — run as Administrator)")
	} else if blocking {
		domains, _ := blockerservice.GetBlockedDomainsFromHosts()
		fmt.Printf("Blocking:  ✓ ACTIVE (%d domains in hosts file)\n", len(domains))
	} else {
		fmt.Println("Blocking:  ✗ INACTIVE (no entries in hosts file)")
	}
}

func cmdSetPassword() {
	hashPath, err := config.PasswordHashPath()
	must(err)
	must(config.EnsureAppDataDir())

	// If a hash already exists, verify old password first
	if _, err := os.Stat(hashPath); err == nil {
		fmt.Print("Enter current password: ")
		old, err := readPassword()
		must(err)
		fmt.Println()
		storedHash, err := os.ReadFile(hashPath)
		must(err)
		if bcrypt.CompareHashAndPassword(storedHash, []byte(old)) != nil {
			fatalf("Incorrect password.")
		}
	}

	fmt.Print("Enter new password: ")
	pw1, err := readPassword()
	must(err)
	fmt.Println()
	if len(pw1) < 4 {
		fatalf("Password must be at least 4 characters.")
	}
	fmt.Print("Confirm new password: ")
	pw2, err := readPassword()
	must(err)
	fmt.Println()
	if pw1 != pw2 {
		fatalf("Passwords do not match.")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pw1), bcrypt.DefaultCost)
	must(err)
	must(os.WriteFile(hashPath, hash, 0600))
	fmt.Println("✓ Password set successfully.")
}

// --- Service action helper ---

func runServiceAction(action string) {
	requireAdmin()
	svc, err := blockerservice.NewService()
	must(err)

	switch action {
	case "install":
		must(svc.Install())
		fmt.Println("✓ Service installed. Run 'start' to activate.")
	case "uninstall":
		must(svc.Uninstall())
		fmt.Println("✓ Service uninstalled.")
	case "start":
		must(svc.Start())
		fmt.Println("✓ Service started. Blocking is now active.")
	case "stop":
		must(svc.Stop())
		fmt.Println("✓ Service stopped. Hosts file entries removed.")
	}
}

// --- Password Helpers ---

// requireAdmin checks that the current process has Administrator privileges.
// If not, it prints a clear error and exits.
func requireAdmin() {
	if !config.IsAdmin() {
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Error: Administrator privileges required.")
		fmt.Fprintln(os.Stderr, "  Please right-click your terminal and choose 'Run as administrator', then try again.")
		fmt.Fprintln(os.Stderr, "")
		os.Exit(1)
	}
}

// requirePassword prompts for a password and verifies it against the stored hash.
// If no password is set yet, prompts the user to set one first.
func requirePassword() {
	hashPath, err := config.PasswordHashPath()
	must(err)

	if _, err := os.Stat(hashPath); os.IsNotExist(err) {
		fmt.Println("No password set yet. Please set a password to protect this tool.")
		cmdSetPassword()
		return
	}

	storedHash, err := os.ReadFile(hashPath)
	must(err)

	fmt.Print("Password: ")
	pw, err := readPassword()
	must(err)
	fmt.Println()

	if bcrypt.CompareHashAndPassword(storedHash, []byte(pw)) != nil {
		fatalf("Incorrect password.")
	}
}

// readPassword reads a password from stdin without echoing characters.
func readPassword() (string, error) {
	// Try to use terminal raw mode for hidden input
	if term.IsTerminal(int(syscall.Stdin)) {
		pw, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", err
		}
		return string(pw), nil
	}
	// Fallback for non-terminal stdin
	reader := bufio.NewReader(os.Stdin)
	pw, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(pw, "\r\n"), nil
}

// --- Utilities ---

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func fatalf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func printHelp() {
	fmt.Print(`
URL Blocker - Parental Control Tool (Windows)
=============================================
Usage: urlblocker <command> [arguments]

Commands (no privileges required):
  status              Show service status and active blocks
  list                List all domains in the blocklist
  group-list          List all mutual-block groups and their live active site
  flush               Flush Windows DNS cache to apply blocks instantly

Commands (password required):
  add <domain>        Add a domain to the blocklist
  remove <domain>     Remove a domain from the blocklist

Mutual-block groups — automatic mode (password required):
  group-add  <name> <domain1> <domain2> [...]
                      Create a group. When the service runs, the FIRST site a
                      user opens gets the lock; all others are blocked until
                      the session expires (default 10 sec idle).
  group-remove <name> Delete a group (domains stay in blocklist.txt)
  group-timeout <name> <seconds>
                      Set the idle-inactivity timeout for a group.

Parent overrides (password + Administrator required):
  group-allow  <name> <domain>
                      Force a specific site to be the active one, bypassing
                      auto-detection. Other group members are blocked.
  group-reset  <name> Return the group to automatic mode (all blocked until
                      one is opened by the user).

Commands (Administrator required):
  disable             Remove all managed hosts file entries

Commands (Administrator + password required):
  enable              Apply the blocklist to the hosts file now

Service management (Administrator required):
  install             Install as a Windows background service
  uninstall           Remove the Windows service
  start               Start the background service
  stop                Stop the background service

Password management:
  setpassword         Set the CLI password (if not already set)
  changepassword      Change the existing CLI password

Examples:
  group-add gaming roblox.com youtube.com twitch.tv
  group-list                      (see who has the lock right now)
  group-timeout gaming 30         (release lock after 30 sec idle)
  group-allow gaming youtube.com  (parent override)
  group-reset gaming              (back to auto mode)
  add facebook.com
  enable
  install
  status

Tip: Run your terminal as Administrator to use enable, disable, group-allow,
     group-reset, and service management commands.
`)
}
