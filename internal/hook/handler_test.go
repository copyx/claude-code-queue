package hook_test

import (
	"testing"

	"github.com/jingikim/ccq/internal/hook"
	"github.com/jingikim/ccq/internal/queue"
	"github.com/jingikim/ccq/internal/switcher"
	"github.com/jingikim/ccq/internal/tmux"
)

func setup(t *testing.T, name string) (*tmux.Tmux, *queue.Queue, *switcher.Switcher, func()) {
	t.Helper()
	if !tmux.IsInstalled() {
		t.Skip("tmux not installed")
	}
	tm := tmux.New(name)
	if err := tm.NewSession(); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	q := queue.New(tm)
	sw := switcher.New(tm, q)
	sw.SetAutoSwitch(true)
	return tm, q, sw, func() { tm.KillSession() }
}

func TestHandleIdle_MarksWindowIdle(t *testing.T) {
	tm, q, sw, cleanup := setup(t, "ccq-test-hook-idle")
	defer cleanup()

	windows, _ := tm.ListWindows()
	w0 := windows[0].ID

	h := hook.New(tm, q, sw)
	if err := h.HandleIdle(w0); err != nil {
		t.Fatalf("HandleIdle: %v", err)
	}

	if !q.IsIdle(w0) {
		t.Error("window should be idle after HandleIdle")
	}
}

func TestHandleBusy_MarksWindowBusyAndSwitches(t *testing.T) {
	tm, q, sw, cleanup := setup(t, "ccq-test-hook-busy")
	defer cleanup()

	windows, _ := tm.ListWindows()
	w0 := windows[0].ID
	w1, _ := tm.NewWindow("/tmp")

	// w0 = active + will become busy, w1 = idle
	q.MarkIdle(w1)

	h := hook.New(tm, q, sw)
	if err := h.HandleBusy(w0); err != nil {
		t.Fatalf("HandleBusy: %v", err)
	}

	if q.IsIdle(w0) {
		t.Error("w0 should be busy")
	}

	activeID, _ := tm.ActiveWindowID()
	if activeID != w1 {
		t.Errorf("expected switch to %s, got active=%s", w1, activeID)
	}
}
