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

	// PasswordFilename is the bcrypt hash file name
	PasswordFilename = "password.hash"

	// AppDirName is the app data folder name
	AppDirName = "urlblocker"

	// PollInterval is how often the service checks for blocklist changes (seconds)
	PollInterval = 30
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
