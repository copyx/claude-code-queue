// internal/queue/queue_test.go
package queue_test

import (
	"testing"
	"time"

	"github.com/jingikim/ccq/internal/queue"
	"github.com/jingikim/ccq/internal/tmux"
)

func TestMarkAndFindOldestIdle(t *testing.T) {
	if !tmux.IsInstalled() {
		t.Skip("tmux not installed")
	}

	tm := tmux.New("ccq-test-queue")
	if err := tm.NewSession(); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer tm.KillSession()

	q := queue.New(tm)

	// Get default window (index 0)
	windows, _ := tm.ListWindows()
	w0 := windows[0].ID

	// Add a second window
	w1, _ := tm.NewWindow("/tmp")

	// Mark both idle (w0 first, w1 later)
	q.MarkIdle(w0)
	time.Sleep(time.Second) // ensure different unix timestamps
	q.MarkIdle(w1)

	// Oldest idle should be w0
	oldest, err := q.OldestIdle()
	if err != nil {
		t.Fatalf("OldestIdle: %v", err)
	}
	if oldest != w0 {
		t.Errorf("expected oldest idle = %s, got %s", w0, oldest)
	}

	// Switch w0 to busy
	q.MarkBusy(w0)

	// Now oldest idle should be w1
	oldest, err = q.OldestIdle()
	if err != nil {
		t.Fatalf("OldestIdle: %v", err)
	}
	if oldest != w1 {
		t.Errorf("expected oldest idle = %s, got %s", w1, oldest)
	}
}

func TestOldestIdle_NoneIdle(t *testing.T) {
	if !tmux.IsInstalled() {
		t.Skip("tmux not installed")
	}

	tm := tmux.New("ccq-test-queue-none")
	if err := tm.NewSession(); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer tm.KillSession()

	q := queue.New(tm)
	windows, _ := tm.ListWindows()
	q.MarkBusy(windows[0].ID)

	oldest, err := q.OldestIdle()
	if err != nil {
		t.Fatalf("OldestIdle: %v", err)
	}
	if oldest != "" {
		t.Errorf("expected no idle window, got %s", oldest)
	}
}
