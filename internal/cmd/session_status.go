package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jingikim/ccq/internal/queue"
	"github.com/jingikim/ccq/internal/tmux"
)

// SessionStatus prints a detailed view of the ccq session for the terminal.
func SessionStatus() error {
	tm := tmux.New(sessionName)
	output, err := renderSessionStatus(tm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(output)
	return nil
}

func renderSessionStatus(tm *tmux.Tmux) (string, error) {
	if !tm.HasSession() {
		return "", fmt.Errorf("ccq: no active session")
	}

	windows, err := tm.ListWindows()
	if err != nil {
		return "", err
	}

	clients := tm.ListClients()
	autoSwitch, _ := tm.GetSessionOption("@ccq_auto_switch")

	// Header
	var b strings.Builder

	windowWord := "windows"
	if len(windows) == 1 {
		windowWord = "window"
	}
	clientWord := "clients"
	if len(clients) == 1 {
		clientWord = "client"
	}

	switchState := "off"
	if autoSwitch == "on" {
		switchState = "on"
	}

	fmt.Fprintf(&b, "ccq: %d %s, %d %s attached, auto-switch %s\n",
		len(windows), windowWord, len(clients), clientWord, switchState)

	if len(windows) == 0 {
		return b.String(), nil
	}

	b.WriteString("\n")

	// Per-window lines
	for _, w := range windows {
		state, _ := tm.GetWindowOption(w.ID, queue.StateKey)
		dir, _ := tm.GetWindowPanePath(w.ID)

		// Shorten home directory
		if home, err := os.UserHomeDir(); err == nil {
			dir = strings.Replace(dir, home, "~", 1)
		}
		if dir == "" {
			dir = "~"
		}

		name := filepath.Base(dir)
		if name == "~" || name == "." || name == "" {
			name = "~"
		}

		stateStr := state
		if stateStr == "" {
			stateStr = "-"
		}

		idleStr := ""
		if state == "idle" {
			sinceStr, _ := tm.GetWindowOption(w.ID, queue.IdleSinceKey)
			if ts, err := strconv.ParseInt(sinceStr, 10, 64); err == nil && ts > 0 {
				idleStr = formatDuration(time.Since(time.Unix(ts, 0)))
			}
		}

		fmt.Fprintf(&b, "  #%-3s %-15s %-6s %6s   %s\n",
			w.Index, name, stateStr, idleStr, dir)
	}

	return b.String(), nil
}
