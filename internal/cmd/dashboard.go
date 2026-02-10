package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jingikim/ccq/internal/config"
	"github.com/jingikim/ccq/internal/dashboard"
	"github.com/jingikim/ccq/internal/queue"
	"github.com/jingikim/ccq/internal/tmux"
)

// Dashboard runs the dashboard display in a loop.
func Dashboard(width int, refreshInterval time.Duration) error {
	tm := tmux.New(sessionName)
	if !tm.HasSession() {
		return fmt.Errorf("session %q not found", sessionName)
	}

	// Load config for dashboard settings
	cfg, err := config.Load(config.DefaultPath())
	if err == nil && cfg.Dashboard.Width > 0 {
		width = cfg.Dashboard.Width
	}
	if width <= 0 {
		width = 20
	}

	if refreshInterval <= 0 {
		refreshInterval = 2 * time.Second
	}

	q := queue.New(tm)
	dash := dashboard.New(tm, q, width)

	// Handle signals for clean shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	// Clear screen and render immediately
	clearScreen()
	if err := renderDashboard(dash); err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			clearScreen()
			if err := renderDashboard(dash); err != nil {
				// Don't exit on render errors, just show the error
				fmt.Printf("Error: %v\n", err)
				// If session is gone, exit gracefully
				if !tm.HasSession() {
					return nil
				}
			}
		case <-sigCh:
			return nil
		}
	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func renderDashboard(dash *dashboard.Dashboard) error {
	output, err := dash.Render()
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
