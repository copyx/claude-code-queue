// Package hook provides handlers for Claude Code hook events (idle, busy, remove).
package hook

import (
	"strings"

	"github.com/jingikim/ccq/internal/queue"
	"github.com/jingikim/ccq/internal/switcher"
	"github.com/jingikim/ccq/internal/tmux"
)

// Handler processes hook events from Claude Code.
type Handler struct {
	tm *tmux.Tmux
	q  *queue.Queue
	sw *switcher.Switcher
}

// New creates a Handler with the given tmux session, queue, and switcher.
func New(tm *tmux.Tmux, q *queue.Queue, sw *switcher.Switcher) *Handler {
	return &Handler{tm: tm, q: q, sw: sw}
}

// HandleIdle marks a window as idle and attempts an auto-switch.
// If the window has @ccq_return_to set (initial setup after ccq add),
// it switches back to the previous window or detaches the client instead.
func (h *Handler) HandleIdle(windowID string) error {
	returnTo, _ := h.tm.GetWindowOption(windowID, "@ccq_return_to")
	if returnTo != "" {
		h.tm.UnsetWindowOption(windowID, "@ccq_return_to")
		if err := h.q.MarkIdle(windowID); err != nil {
			return err
		}
		if returnTo == "__detach__" {
			h.tm.Run("detach-client", "-s", h.tm.Session)
		} else if strings.HasPrefix(returnTo, "__detach__:") {
			tty := strings.TrimPrefix(returnTo, "__detach__:")
			h.tm.Run("detach-client", "-t", tty)
		} else {
			h.tm.SelectWindow(returnTo)
		}
		return nil
	}

	if err := h.q.MarkIdle(windowID); err != nil {
		return err
	}
	h.sw.TrySwitch()
	return nil
}

// HandleBusy marks a window as busy and attempts an auto-switch.
// If the window is already idle, skip marking as busy to prevent race conditions
// where PostToolUse fires after a Notification hook has already set the window to idle.
func (h *Handler) HandleBusy(windowID string) error {
	// Don't override idle state - it's more recent and accurate
	if h.q.IsIdle(windowID) {
		return nil
	}

	if err := h.q.MarkBusy(windowID); err != nil {
		return err
	}
	h.sw.TrySwitch()
	return nil
}

// HandlePromptSubmit marks a window as busy (overriding idle state) and attempts auto-switch.
// Used for UserPromptSubmit hook - when user submits a prompt, the window transitions
// from idle to busy, so we should always mark as busy and switch.
func (h *Handler) HandlePromptSubmit(windowID string) error {
	if err := h.q.MarkBusy(windowID); err != nil {
		return err
	}
	h.sw.TrySwitch()
	return nil
}

// HandleRemove clears all ccq-related window options for a removed window.
// Uses UnsetWindowOption to properly remove variables. Errors are ignored
// because the window may already be gone (remain-on-exit off).
func (h *Handler) HandleRemove(windowID string) error {
	_ = h.tm.UnsetWindowOption(windowID, "@ccq_state")
	_ = h.tm.UnsetWindowOption(windowID, "@ccq_idle_since")
	return nil
}
