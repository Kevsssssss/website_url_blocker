package service

import (
	"os/exec"
	"strings"
)

// redirectDNS sets the DNS server for all connected network adapters to
// 127.0.0.1 so every DNS query on this machine is handled by the local proxy.
func redirectDNS() error {
	for _, adapter := range getActiveAdapters() {
		exec.Command(
			"netsh", "interface", "ip", "set", "dns",
			"name="+adapter, "source=static", "address=127.0.0.1", "validate=no",
		).Run()
	}
	return nil
}

// restoreDNS resets DNS back to DHCP (automatic) for all connected adapters,
// reversing the effect of redirectDNS when the service stops.
func restoreDNS() error {
	for _, adapter := range getActiveAdapters() {
		exec.Command(
			"netsh", "interface", "ip", "set", "dns",
			"name="+adapter, "source=dhcp",
		).Run()
	}
	return nil
}

// getActiveAdapters returns the names of currently connected network adapters
// by parsing the output of `netsh interface show interface`.
// Falls back to ["Wi-Fi", "Ethernet"] if parsing fails.
func getActiveAdapters() []string {
	out, err := exec.Command("netsh", "interface", "show", "interface").Output()
	if err != nil {
		return []string{"Wi-Fi", "Ethernet"}
	}

	// Output format:
	// Admin State    State          Type             Interface Name
	// ---------------------------------------------------------------
	// Enabled        Connected      Dedicated        Wi-Fi
	// Enabled        Disconnected   Dedicated        Ethernet
	// Enabled        Connected      Loopback         Loopback Pseudo-Interface 1

	var adapters []string
	for _, line := range strings.Split(string(out), "\n") {
		// Only pick connected, non-loopback adapters.
		if !strings.Contains(line, "Connected") {
			continue
		}
		fields := strings.Fields(line)
		// Need at least 4 fields: AdminState State Type Name...
		if len(fields) < 4 {
			continue
		}
		adapterType := fields[2]
		if strings.EqualFold(adapterType, "Loopback") {
			continue
		}
		// Interface name is everything from field index 3 onward.
		name := strings.Join(fields[3:], " ")
		adapters = append(adapters, name)
	}

	if len(adapters) == 0 {
		return []string{"Wi-Fi", "Ethernet"}
	}
	return adapters
}
