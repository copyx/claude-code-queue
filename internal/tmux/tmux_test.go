// internal/tmux/tmux_test.go
package tmux_test

import (
	"testing"

	"github.com/jingikim/ccq/internal/tmux"
)

func TestHasSession_NoSession(t *testing.T) {
	tm := tmux.New("ccq-test-nonexistent")
	if tm.HasSession() {
		t.Error("expected HasSession() to return false for nonexistent session")
	}
}

func TestSessionLifecycle(t *testing.T) {
	if !tmux.IsInstalled() {
		t.Skip("tmux not installed")
	}

	name := "ccq-test-lifecycle"
	tm := tmux.New(name)

	// Verify no session exists yet
	if tm.HasSession() {
		t.Fatal("session should not exist yet")
	}

	// Create detached session
	if err := tm.NewSession(); err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	defer tm.KillSession()

	// Verify session exists
	if !tm.HasSession() {
		t.Error("session should exist after NewSession")
	}

	// Kill session
	if err := tm.KillSession(); err != nil {
		t.Fatalf("KillSession failed: %v", err)
	}
	if tm.HasSession() {
		t.Error("session should not exist after KillSession")
	}
}

func TestWindowAndOptions(t *testing.T) {
	if !tmux.IsInstalled() {
		t.Skip("tmux not installed")
	}

	name := "ccq-test-window"
	tm := tmux.New(name)
	if err := tm.NewSession(); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer tm.KillSession()

	// Add a new window
	windowID, err := tm.NewWindow("/tmp")
	if err != nil {
		t.Fatalf("NewWindow: %v", err)
	}

	// List windows
	windows, err := tm.ListWindows()
	if err != nil {
		t.Fatalf("ListWindows: %v", err)
	}
	if len(windows) < 2 {
		t.Errorf("expected at least 2 windows, got %d", len(windows))
	}

	// Set/get window option
	if err := tm.SetWindowOption(windowID, "@ccq_state", "idle"); err != nil {
		t.Fatalf("SetWindowOption: %v", err)
	}
	val, err := tm.GetWindowOption(windowID, "@ccq_state")
	if err != nil {
		t.Fatalf("GetWindowOption: %v", err)
	}
	if val != "idle" {
		t.Errorf("expected 'idle', got %q", val)
	}
}

func TestListClients(t *testing.T) {
	if !tmux.IsInstalled() {
		t.Skip("tmux not installed")
	}

	name := "ccq-test-clients"
	tm := tmux.New(name)
	if err := tm.NewSession(); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer tm.KillSession()

	// Detached session should have no clients
	clients := tm.ListClients()
	if len(clients) != 0 {
		t.Errorf("expected 0 clients for detached session, got %d", len(clients))
	}
}
