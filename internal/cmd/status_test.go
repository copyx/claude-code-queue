package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/jingikim/ccq/internal/queue"
	"github.com/jingikim/ccq/internal/tmux"
)

func TestRenderStatusLine(t *testing.T) {
	if !tmux.IsInstalled() {
		t.Skip("tmux not installed")
	}

	tm := tmux.New("ccq-test-status")
	if err := tm.NewSession(); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer tm.KillSession()

	q := queue.New(tm)

	// Mark first window as idle
	windows, _ := tm.ListWindows()
	if len(windows) > 0 {
		q.MarkIdle(windows[0].ID)
	}

	// Add a second window and mark it busy
	w1, err := tm.NewWindow("/tmp")
	if err != nil {
		t.Fatalf("NewWindow: %v", err)
	}
	q.MarkBusy(w1)

	// Add a third window (active) so window 0 and 1 show their real states
	w2, err := tm.NewWindow("/tmp")
	if err != nil {
		t.Fatalf("NewWindow: %v", err)
	}
	tm.SelectWindow(w2)

	time.Sleep(100 * time.Millisecond)

	line, err := renderStatusLine(tm)
	if err != nil {
		t.Fatalf("renderStatusLine: %v", err)
	}

	// Should be a single line (no newlines)
	if strings.Contains(line, "\n") {
		t.Error("status line should not contain newlines")
	}

	// Should contain window indices
	if !strings.Contains(line, "0:") {
		t.Error("status line should contain window 0")
	}
	if !strings.Contains(line, "1:") {
		t.Error("status line should contain window 1")
	}

	// Should contain idle icon for first window
	if !strings.Contains(line, "○") {
		t.Error("status line should contain idle icon ○")
	}

	// Should contain busy icon for second window
	if !strings.Contains(line, "●") {
		t.Error("status line should contain busy icon ●")
	}

	// Should contain active icon for third window
	if !strings.Contains(line, "▶") {
		t.Error("status line should contain active icon ▶")
	}

	// Should contain summary (only window 0 is idle)
	if !strings.Contains(line, "1/3 idle") {
		t.Errorf("status line should contain '1/3 idle', got: %s", line)
	}
}

func TestRenderStatusLineActiveWindow(t *testing.T) {
	if !tmux.IsInstalled() {
		t.Skip("tmux not installed")
	}

	tm := tmux.New("ccq-test-status-active")
	if err := tm.NewSession(); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer tm.KillSession()

	line, err := renderStatusLine(tm)
	if err != nil {
		t.Fatalf("renderStatusLine: %v", err)
	}

	// Active window should have ▶ icon
	if !strings.Contains(line, "▶") {
		t.Error("status line should contain active icon ▶")
	}

	// Summary should show 0 idle
	if !strings.Contains(line, "0/1 idle") {
		t.Errorf("status line should contain '0/1 idle', got: %s", line)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m"},
		{5 * time.Minute, "5m"},
		{2 * time.Hour, "2h"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
