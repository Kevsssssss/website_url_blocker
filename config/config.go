package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	// HostsFilePath is the Windows hosts file location
	HostsFilePath = `C:\Windows\System32\drivers\etc\hosts`

	// ServiceName is the Windows service identifier
	ServiceName = "URLBlocker"

	// ServiceDisplayName is the human-readable name in services.msc
	ServiceDisplayName = "URL Blocker - Parental Control"

	// ServiceDescription is shown in services.msc
	ServiceDescription = "Blocks domains listed in blocklist.txt by modifying the Windows hosts file. Lightweight parental control tool."

	// BlocklistFilename is the name of the domain blocklist file
	BlocklistFilename = "blocklist.txt"

	// GroupsFilename is the name of the mutual-block groups config file
	GroupsFilename = "groups.json"

	// GroupsStateFilename stores the live DNS-proxy session state (read by group-list)
	GroupsStateFilename = "groups_state.json"

	// PasswordFilename is the bcrypt hash file name
	PasswordFilename = "password.hash"

	// AppDirName is the app data folder name
	AppDirName = "urlblocker"

	// PollInterval is how often the service checks for blocklist changes (seconds)
	PollInterval = 30

	// DNSListenAddr is the address the local DNS proxy listens on
	DNSListenAddr = "127.0.0.1:53"

	// DNSUpstream is the real DNS server queries are forwarded to
	DNSUpstream = "1.1.1.1:53"

	// GroupSessionTimeout is the default inactivity timeout in minutes before
	// the active site lock in a group is automatically released
	GroupSessionTimeout = 5
)

// AppDataDir returns the path to the app's data directory (%APPDATA%\urlblocker)
func AppDataDir() (string, error) {
	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("this tool only supports Windows")
	}
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("APPDATA environment variable not set")
	}
	return filepath.Join(appData, AppDirName), nil
}

// PasswordHashPath returns the full path to the password hash file
func PasswordHashPath() (string, error) {
	dir, err := AppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, PasswordFilename), nil
}

// EnsureAppDataDir creates the app data directory if it doesn't exist
func EnsureAppDataDir() error {
	dir, err := AppDataDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0700)
}

// BlocklistPath returns the path to blocklist.txt (same dir as the executable)
func BlocklistPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not determine executable path: %w", err)
	}
	return filepath.Join(filepath.Dir(exe), BlocklistFilename), nil
}

// GroupsPath returns the path to groups.json (same dir as the executable)
func GroupsPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not determine executable path: %w", err)
	}
	return filepath.Join(filepath.Dir(exe), GroupsFilename), nil
}

// GroupsStatePath returns the path to groups_state.json, which the DNS proxy
// writes with the live session state so the CLI can display it in group-list.
func GroupsStatePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not determine executable path: %w", err)
	}
	return filepath.Join(filepath.Dir(exe), GroupsStateFilename), nil
}
