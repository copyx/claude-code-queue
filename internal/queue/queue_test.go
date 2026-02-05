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

	// 기본 window (index 0)의 ID
	windows, _ := tm.ListWindows()
	w0 := windows[0].ID

	// 두 번째 window 추가
	w1, _ := tm.NewWindow("/tmp")

	// 둘 다 idle로 마킹 (w0 먼저, w1 나중)
	q.MarkIdle(w0)
	time.Sleep(time.Second) // ensure different unix timestamps
	q.MarkIdle(w1)

	// 가장 오래된 idle = w0
	oldest, err := q.OldestIdle()
	if err != nil {
		t.Fatalf("OldestIdle: %v", err)
	}
	if oldest != w0 {
		t.Errorf("expected oldest idle = %s, got %s", w0, oldest)
	}

	// w0을 busy로 전환
	q.MarkBusy(w0)

	// 이제 가장 오래된 idle = w1
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
