package service

import "os"

// statFile is a thin wrapper around os.Stat for easy testing/mocking.
func statFile(path string) (os.FileInfo, error) {
	return os.Stat(path)
}
