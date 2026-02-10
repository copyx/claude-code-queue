package cmd

import (
	"fmt"
	"os"

	"github.com/jingikim/ccq/internal/hook"
	"github.com/jingikim/ccq/internal/queue"
	"github.com/jingikim/ccq/internal/switcher"
	"github.com/jingikim/ccq/internal/tmux"
)

func Hook(action string) error {
	pane := os.Getenv("TMUX_PANE")
	if pane == "" {
		return fmt.Errorf("TMUX_PANE not set (not running inside tmux?)")
	}

	tm := tmux.New(sessionName)
	if !tm.HasSession() {
		return nil
	}

	windowID, err := tm.WindowIDFromPane(pane)
	if err != nil {
		return fmt.Errorf("failed to resolve window from pane %s: %w", pane, err)
	}

	q := queue.New(tm)
	sw := switcher.New(tm, q)
	h := hook.New(tm, q, sw)

	switch action {
	case "idle":
		return h.HandleIdle(windowID)
	case "busy":
		return h.HandleBusy(windowID)
	case "prompt":
		return h.HandlePromptSubmit(windowID)
	case "remove":
		return h.HandleRemove(windowID)
	default:
		return fmt.Errorf("unknown hook action: %s", action)
	}
}
