package main

import (
	"fmt"
	"os"

	"github.com/jingikim/ccq/internal/cmd"
)

var Version = "dev" // overwritten by ldflags at build time

func printHelp() {
	fmt.Print(`ccq - Claude Code Queue Manager

FIFO queue-based auto-switcher for multiple Claude Code sessions via tmux.

Usage:
  ccq             Start ccq or add a new Claude window
  ccq attach      Attach to existing session (no new window)
  ccq status      Show session status
  ccq -h, --help  Show this help
  ccq --version   Show version

Keybindings (inside ccq session):
  prefix + a      Toggle auto/manual switching
  prefix + g      Toggle dashboard (gauge)
  prefix + n/p    Next/previous window
  prefix + w      Window list
  prefix + d      Detach from session
`)
}

func main() {
	var err error

	if len(os.Args) < 2 {
		err = cmd.Root()
	} else {
		switch os.Args[1] {
		case "-h", "--help", "help":
			printHelp()
		case "--version", "-v":
			fmt.Printf("ccq version %s\n", Version)
			return
		case "_hook":
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "usage: ccq _hook <idle|busy|prompt|remove>")
				os.Exit(1)
			}
			err = cmd.Hook(os.Args[2])
		case "_toggle":
			err = cmd.Toggle()
		case "_status":
			err = cmd.Status()
		case "status":
			err = cmd.SessionStatus()
		case "attach":
			err = cmd.Attach()
		case "toggle-dashboard":
			err = cmd.ToggleDashboard()
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
			fmt.Fprintln(os.Stderr, "Run 'ccq -h' for usage.")
			os.Exit(1)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
