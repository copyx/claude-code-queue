package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/jingikim/ccq/internal/config"
	"github.com/jingikim/ccq/internal/tmux"
)

const dashboardPaneKey = "@ccq_dashboard_pane"

// ToggleDashboard shows or hides the dashboard pane.
func ToggleDashboard() error {
	tm := tmux.New(sessionName)
	if !tm.HasSession() {
		return fmt.Errorf("session %q not found", sessionName)
	}

	// Check if dashboard pane exists
	paneID, _ := tm.GetSessionOption(dashboardPaneKey)
	if paneID != "" && paneExists(paneID) {
		// Dashboard exists, kill it
		if err := killPane(paneID); err != nil {
			return fmt.Errorf("failed to kill dashboard pane: %w", err)
		}
		tm.SetSessionOption(dashboardPaneKey, "")
		return nil
	}

	// Dashboard doesn't exist, create it
	return createDashboard(tm)
}

func paneExists(paneID string) bool {
	cmd := exec.Command("tmux", "display-message", "-t", paneID, "-p", "#{pane_id}")
	return cmd.Run() == nil
}

func killPane(paneID string) error {
	cmd := exec.Command("tmux", "kill-pane", "-t", paneID)
	return cmd.Run()
}

func createDashboard(tm *tmux.Tmux) error {
	// Load config for dashboard width
	width := 20 // Default width
	cfg, err := config.Load(config.DefaultPath())
	if err == nil && cfg.Dashboard.Width > 0 {
		width = cfg.Dashboard.Width
	}

	// Split window horizontally, creating a pane on the right
	// -h: horizontal split, -l: size (columns)
	// -d: don't switch to new pane
	// -P -F: print format (get pane ID)
	out, err := tm.Run("split-window", "-h", "-l", fmt.Sprintf("%d", width), "-d",
		"-P", "-F", "#{pane_id}",
		"ccq", "dashboard")
	if err != nil {
		return fmt.Errorf("failed to create dashboard pane: %w", err)
	}

	paneID := strings.TrimSpace(out)
	if paneID == "" {
		return fmt.Errorf("failed to get dashboard pane ID")
	}

	// Store pane ID in session option
	if err := tm.SetSessionOption(dashboardPaneKey, paneID); err != nil {
		return fmt.Errorf("failed to store dashboard pane ID: %w", err)
	}

	return nil
}
