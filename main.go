package main

import (
	"os"

	"github.com/Kevsssssss/website_url_blocker/cli"
	"github.com/Kevsssssss/website_url_blocker/service"
	svc "github.com/kardianos/service"
)

func main() {
	// If no args OR first arg is a CLI command, run CLI mode
	args := os.Args[1:]

	// kardianos/service passes a special flag when run as a service
	// We detect this by checking if the service manager invoked us
	s, err := service.NewService()
	if err == nil {
		// Check if we're being invoked by the Windows Service Control Manager
		// (no args, or service-specific args)
		if svc.Interactive() {
			// Running interactively — dispatch CLI
			cli.Run(args)
		} else {
			// Running as a Windows service — enter service loop
			if err := s.Run(); err != nil {
				os.Exit(1)
			}
		}
	} else {
		// Fallback: just run CLI if service can't be created
		cli.Run(args)
	}
}
