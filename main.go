package main

import (
	"os"

	"github.com/Kevsssssss/website_url_blocker/cli"
	"github.com/Kevsssssss/website_url_blocker/repl"
	"github.com/Kevsssssss/website_url_blocker/service"
	svc "github.com/kardianos/service"
)

func main() {
	args := os.Args[1:]

	// If any arguments are provided, always dispatch to CLI.
	// Never let svc.Interactive() interfere with explicit commands.
	if len(args) > 0 {
		cli.Run(args)
		return
	}

	// No arguments: check if we were invoked by the Windows Service Control Manager.
	if !svc.Interactive() {
		s, err := service.NewService()
		if err != nil {
			os.Exit(1)
		}
		if err := s.Run(); err != nil {
			os.Exit(1)
		}
		return
	}

	// No arguments, running interactively (double-clicked or run from terminal):
	// launch the interactive REPL shell.
	repl.Start()
}
