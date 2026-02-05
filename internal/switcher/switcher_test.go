package switcher_test

import (
	"testing"

	"github.com/jingikim/ccq/internal/queue"
	"github.com/jingikim/ccq/internal/switcher"
	"github.com/jingikim/ccq/internal/tmux"
)

func setup(t *testing.T, name string) (*tmux.Tmux, *queue.Queue, func()) {
	t.Helper()
	if !tmux.IsInstalled() {
		t.Skip("tmux not installed")
	}
	tm := tmux.New(name)
	if err := tm.NewSession(); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	q := queue.New(tm)
	return tm, q, func() { tm.KillSession() }
}

func TestAutoSwitch_CurrentBusy_SwitchesToOldestIdle(t *testing.T) {
	tm, q, cleanup := setup(t, "ccq-test-switch-1")
	defer cleanup()

	windows, _ := tm.ListWindows()
	w0 := windows[0].ID
	w1, _ := tm.NewWindow("/tmp")

	// w0=busy (currently active), w1=idle
	q.MarkBusy(w0)
	q.MarkIdle(w1)

	sw := switcher.New(tm, q)
	sw.SetAutoSwitch(true)
	switched := sw.TrySwitch()

	if !switched {
		t.Error("expected switch to happen")
	}

	activeID, _ := tm.ActiveWindowID()
	if activeID != w1 {
		t.Errorf("expected active window = %s, got %s", w1, activeID)
	}
}

func TestAutoSwitch_CurrentIdle_NoSwitch(t *testing.T) {
	tm, q, cleanup := setup(t, "ccq-test-switch-2")
	defer cleanup()

	windows, _ := tm.ListWindows()
	w0 := windows[0].ID
	w1, _ := tm.NewWindow("/tmp")

	// both idle, user might be typing in w0
	q.MarkIdle(w0)
	q.MarkIdle(w1)

	sw := switcher.New(tm, q)
	sw.SetAutoSwitch(true)
	switched := sw.TrySwitch()

	if switched {
		t.Error("expected no switch when current window is idle")
	}
}

func TestAutoSwitch_Disabled_NoSwitch(t *testing.T) {
	tm, q, cleanup := setup(t, "ccq-test-switch-3")
	defer cleanup()

	windows, _ := tm.ListWindows()
	w0 := windows[0].ID
	w1, _ := tm.NewWindow("/tmp")

	q.MarkBusy(w0)
	q.MarkIdle(w1)

	sw := switcher.New(tm, q)
	sw.SetAutoSwitch(false) // toggle OFF
	switched := sw.TrySwitch()

	if switched {
		t.Error("expected no switch when auto-switch is disabled")
	}
}
