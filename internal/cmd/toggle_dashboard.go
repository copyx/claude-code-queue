package cmd

import (
	"fmt"

	"github.com/jingikim/ccq/internal/tmux"
)

// ToggleDashboard toggles the dashboard status bar line on/off.
// Switches between status 2 (dashboard visible) and status 1 (hidden).
func ToggleDashboard() error {
	tm := tmux.New(sessionName)
	if !tm.HasSession() {
		return fmt.Errorf("session %q not found", sessionName)
	}

	current, _ := tm.GetSessionOption("status")
	if current == "2" {
		return tm.SetSessionOption("status", "on")
	}
	return tm.SetSessionOption("status", "2")
}
