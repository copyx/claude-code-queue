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

func TestHandleBusy_DoesNotOverrideIdleState(t *testing.T) {
	tm, q, sw, cleanup := setup(t, "ccq-test-busy-idle-race")
	defer cleanup()

	windows, _ := tm.ListWindows()
	w0 := windows[0].ID

	h := hook.New(tm, q, sw)

	// Simulate race condition: window becomes idle (Notification hook),
	// then PostToolUse fires asynchronously and tries to mark as busy
	q.MarkIdle(w0)
	if err := h.HandleBusy(w0); err != nil {
		t.Fatalf("HandleBusy: %v", err)
	}

	// Window should remain idle - don't override more recent idle state
	if !q.IsIdle(w0) {
		t.Error("HandleBusy should not override idle state (prevents race condition)")
	}
}

func TestHandlePromptSubmit_SwitchesFromIdleWindow(t *testing.T) {
	tm, q, sw, cleanup := setup(t, "ccq-test-prompt-submit")
	defer cleanup()

	windows, _ := tm.ListWindows()
	w0 := windows[0].ID
	w1, _ := tm.NewWindow("/tmp")

	// Simulate UserPromptSubmit scenario:
	// 1. w0 = idle (user is at idle_prompt)
	// 2. w1 = also idle
	// 3. User submits prompt in w0 → HandlePromptSubmit(w0) should mark busy and switch to w1
	q.MarkIdle(w0)
	q.MarkIdle(w1)
	tm.SelectWindow(w0)

	h := hook.New(tm, q, sw)
	if err := h.HandlePromptSubmit(w0); err != nil {
		t.Fatalf("HandlePromptSubmit: %v", err)
	}

	// Window should now be busy (override idle state)
	if q.IsIdle(w0) {
		t.Error("w0 should be marked as busy after prompt submit")
	}

	// Should switch to w1 (oldest idle)
	activeID, _ := tm.ActiveWindowID()
	if activeID != w1 {
		t.Errorf("expected switch to %s (oldest idle), got active=%s", w1, activeID)
	}
}

func TestHandleIdle_ReturnToDetach(t *testing.T) {
	tm, q, sw, cleanup := setup(t, "ccq-test-hook-detach")
	defer cleanup()

	windows, _ := tm.ListWindows()
	w0 := windows[0].ID

	// Set return_to = "__detach__" (no tty)
	tm.SetWindowOption(w0, "@ccq_return_to", "__detach__")

	h := hook.New(tm, q, sw)
	if err := h.HandleIdle(w0); err != nil {
		t.Fatalf("HandleIdle: %v", err)
	}

	// Window should still be marked idle
	if !q.IsIdle(w0) {
		t.Error("window should be idle after HandleIdle with __detach__")
	}

	// return_to should be cleared
	val, _ := tm.GetWindowOption(w0, "@ccq_return_to")
	if val != "" {
		t.Errorf("expected @ccq_return_to to be cleared, got %q", val)
	}
}

func TestHandleIdle_ReturnToWindow(t *testing.T) {
	tm, q, sw, cleanup := setup(t, "ccq-test-hook-return")
	defer cleanup()

	windows, _ := tm.ListWindows()
	w0 := windows[0].ID
	w1, _ := tm.NewWindow("/tmp")

	// Set return_to = w0 (return to previous window)
	tm.SetWindowOption(w1, "@ccq_return_to", w0)
	tm.SelectWindow(w1)

	h := hook.New(tm, q, sw)
	if err := h.HandleIdle(w1); err != nil {
		t.Fatalf("HandleIdle: %v", err)
	}

	// w1 should be idle
	if !q.IsIdle(w1) {
		t.Error("window should be idle after HandleIdle with return_to")
	}

	// Should have switched back to w0
	activeID, _ := tm.ActiveWindowID()
	if activeID != w0 {
		t.Errorf("expected switch to %s, got active=%s", w0, activeID)
	}
}

func TestHandleRemove_ClearsWindowOptions(t *testing.T) {
	tm, q, sw, cleanup := setup(t, "ccq-test-remove")
	defer cleanup()

	windows, _ := tm.ListWindows()
	windowID := windows[0].ID

	h := hook.New(tm, q, sw)

	q.MarkIdle(windowID)
	h.HandleRemove(windowID)

	state, _ := tm.GetWindowOption(windowID, "@ccq_state")
	if state != "" {
		t.Errorf("expected @ccq_state to be cleared, got %q", state)
	}
}
