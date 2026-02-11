package cmd

import (
	"strings"
	"testing"

	"github.com/jingikim/ccq/internal/queue"
	"github.com/jingikim/ccq/internal/tmux"
)

func TestRenderSessionStatus(t *testing.T) {
	if !tmux.IsInstalled() {
		t.Skip("tmux not installed")
	}

	tm := tmux.New("ccq-test-session-status")
	if err := tm.NewSession(); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer tm.KillSession()

	tm.SetSessionOption("@ccq_auto_switch", "on")

	windows, _ := tm.ListWindows()
	w0 := windows[0].ID

	// Mark w0 as idle with a known timestamp
	q := queue.New(tm)
	q.MarkIdle(w0)

	output, err := renderSessionStatus(tm)
	if err != nil {
		t.Fatalf("renderSessionStatus: %v", err)
	}

	// Should contain header with window count and auto-switch state
	if !strings.Contains(output, "1 window") {
		t.Errorf("expected '1 window' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "auto-switch on") {
		t.Errorf("expected 'auto-switch on' in output, got:\n%s", output)
	}
	// Should contain window line with state
	if !strings.Contains(output, "idle") {
		t.Errorf("expected 'idle' in output, got:\n%s", output)
	}
}

func TestRenderSessionStatus_NoSession(t *testing.T) {
	if !tmux.IsInstalled() {
		t.Skip("tmux not installed")
	}

	tm := tmux.New("ccq-test-no-session-status")

	_, err := renderSessionStatus(tm)
	if err == nil {
		t.Error("expected error for non-existent session")
	}
}
