// Package switcher implements auto-switch logic that moves the user
// to the oldest idle tmux window when the current window is busy.
package switcher

import (
	"github.com/jingikim/ccq/internal/queue"
	"github.com/jingikim/ccq/internal/tmux"
)

const autoSwitchKey = "@ccq_auto_switch"

// Switcher manages automatic window switching based on queue state.
type Switcher struct {
	tm *tmux.Tmux
	q  *queue.Queue
}

// New creates a Switcher for the given tmux session and queue.
func New(tm *tmux.Tmux, q *queue.Queue) *Switcher {
	return &Switcher{tm: tm, q: q}
}

// SetAutoSwitch enables or disables auto-switch via a session option.
func (s *Switcher) SetAutoSwitch(enabled bool) error {
	val := "off"
	if enabled {
		val = "on"
	}
	return s.tm.SetSessionOption(autoSwitchKey, val)
}

// IsAutoSwitchOn returns true if auto-switch is currently enabled.
func (s *Switcher) IsAutoSwitchOn() bool {
	val, _ := s.tm.GetSessionOption(autoSwitchKey)
	return val == "on"
}

// TrySwitch attempts an auto-switch. Returns true if a switch occurred.
// Rules:
// 1. If auto-switch is off, do not switch.
// 2. If the current window is idle, do not switch (user may be typing).
// 3. If the current window is busy, switch to the oldest idle window.
func (s *Switcher) TrySwitch() bool {
	if !s.IsAutoSwitchOn() {
		return false
	}

	activeID, err := s.tm.ActiveWindowID()
	if err != nil {
		return false
	}

	if s.q.IsIdle(activeID) {
		return false
	}

	target, err := s.q.OldestIdle()
	if err != nil || target == "" {
		return false
	}

	if err := s.tm.SelectWindow(target); err != nil {
		return false
	}
	return true
}
