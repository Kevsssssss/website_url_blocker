package repl

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Start launches the interactive REPL shell.
// It is entered when urlblocker.exe is run with no arguments.
func Start() {
	printBanner()

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error resolving executable path:", err)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("urlblocker> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" || line == "q" {
			fmt.Println("\nGoodbye!")
			break
		}

		args := strings.Fields(line)
		runCmd(exe, args)
		fmt.Println()
	}
}

// runCmd re-launches the same executable with the given args as a subprocess.
// This way os.Exit calls inside CLI commands don't kill the shell.
func runCmd(exe string, args []string) {
	cmd := exec.Command(exe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func printBanner() {
	fmt.Println()
	fmt.Println("  +--------------------------------------+")
	fmt.Println("  |   URL Blocker - Parental Control    |")
	fmt.Println("  |         Interactive Shell           |")
	fmt.Println("  +--------------------------------------+")
	fmt.Println()
	fmt.Println("  Type 'help'        -> see all commands")
	fmt.Println("  Type 'exit'        -> close this window")
	fmt.Println()
}
