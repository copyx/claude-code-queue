// Package hook provides handlers for Claude Code hook events (idle, busy, remove).
package hook

import (
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
func (h *Handler) HandleIdle(windowID string) error {
	if err := h.q.MarkIdle(windowID); err != nil {
		return err
	}
	h.sw.TrySwitch()
	return nil
}

// HandleBusy marks a window as busy and attempts an auto-switch.
func (h *Handler) HandleBusy(windowID string) error {
	if err := h.q.MarkBusy(windowID); err != nil {
		return err
	}
	h.sw.TrySwitch()
	return nil
}

// HandleRemove clears all ccq-related window options for a removed window.
func (h *Handler) HandleRemove(windowID string) error {
	_ = h.tm.SetWindowOption(windowID, "@ccq_state", "")
	_ = h.tm.SetWindowOption(windowID, "@ccq_idle_since", "")
	return nil
}
