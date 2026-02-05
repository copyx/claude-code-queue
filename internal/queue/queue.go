// Package queue manages window state (idle/busy) and finds the oldest idle window.
package queue

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jingikim/ccq/internal/tmux"
)

const (
	stateKey     = "@ccq_state"
	idleSinceKey = "@ccq_idle_since"
)

// Queue tracks tmux window states using window-level options.
type Queue struct {
	tm *tmux.Tmux
}

// New creates a Queue for the given tmux session.
func New(tm *tmux.Tmux) *Queue {
	return &Queue{tm: tm}
}

// MarkIdle marks a window as idle and records the current timestamp.
func (q *Queue) MarkIdle(windowID string) error {
	now := fmt.Sprintf("%d", time.Now().Unix())
	if err := q.tm.SetWindowOption(windowID, stateKey, "idle"); err != nil {
		return err
	}
	return q.tm.SetWindowOption(windowID, idleSinceKey, now)
}

// MarkBusy marks a window as busy and clears the idle timestamp.
func (q *Queue) MarkBusy(windowID string) error {
	if err := q.tm.SetWindowOption(windowID, stateKey, "busy"); err != nil {
		return err
	}
	return q.tm.SetWindowOption(windowID, idleSinceKey, "0")
}

// OldestIdle returns the window ID that has been idle the longest.
// Returns "" if no window is idle.
func (q *Queue) OldestIdle() (string, error) {
	windows, err := q.tm.ListWindows()
	if err != nil {
		return "", err
	}

	var oldestID string
	var oldestTime int64 = 1<<63 - 1

	for _, w := range windows {
		state, _ := q.tm.GetWindowOption(w.ID, stateKey)
		if state != "idle" {
			continue
		}
		sinceStr, _ := q.tm.GetWindowOption(w.ID, idleSinceKey)
		since, err := strconv.ParseInt(sinceStr, 10, 64)
		if err != nil || since <= 0 {
			continue
		}
		if since < oldestTime {
			oldestTime = since
			oldestID = w.ID
		}
	}
	return oldestID, nil
}

// IsIdle returns true if the window is currently marked idle.
func (q *Queue) IsIdle(windowID string) bool {
	state, _ := q.tm.GetWindowOption(windowID, stateKey)
	return state == "idle"
}
