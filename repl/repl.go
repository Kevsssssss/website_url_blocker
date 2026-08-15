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
	fmt.Println("  \u250c\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2510")
	fmt.Println("  \u2502   \U0001f6e1\ufe0f  URL Blocker - Parental Control  \u2502")
	fmt.Println("  \u2502          Interactive Shell           \u2502")
	fmt.Println("  \u2514\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2518")
	fmt.Println()
	fmt.Println("  Type 'help'        \u2192 see all commands")
	fmt.Println("  Type 'exit'        \u2192 close this window")
	fmt.Println()
}
