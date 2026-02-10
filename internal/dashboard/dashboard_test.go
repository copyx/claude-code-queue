package dashboard_test

import (
	"strings"
	"testing"
	"time"

	"github.com/jingikim/ccq/internal/dashboard"
	"github.com/jingikim/ccq/internal/queue"
	"github.com/jingikim/ccq/internal/tmux"
)

func TestDashboardRender(t *testing.T) {
	if !tmux.IsInstalled() {
		t.Skip("tmux not installed")
	}

	tm := tmux.New("ccq-test-dashboard")
	if err := tm.NewSession(); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer tm.KillSession()

	q := queue.New(tm)
	dash := dashboard.New(tm, q, 20)

	// Mark first window as idle
	windows, _ := tm.ListWindows()
	if len(windows) > 0 {
		q.MarkIdle(windows[0].ID)
	}

	// Wait a moment for state to propagate
	time.Sleep(100 * time.Millisecond)

	output, err := dash.Render()
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// Check output contains expected elements
	if !strings.Contains(output, "CCQ Dashboard") {
		t.Error("output should contain dashboard title")
	}
	if !strings.Contains(output, "W0") {
		t.Error("output should contain window index")
	}
	// Check for border characters
	if !strings.Contains(output, "┌") || !strings.Contains(output, "└") {
		t.Error("output should contain border characters")
	}
}

func TestDashboardCustomWidth(t *testing.T) {
	if !tmux.IsInstalled() {
		t.Skip("tmux not installed")
	}

	tm := tmux.New("ccq-test-dashboard-width")
	if err := tm.NewSession(); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer tm.KillSession()

	q := queue.New(tm)
	dash := dashboard.New(tm, q, 30)

	output, err := dash.Render()
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	lines := strings.Split(output, "\n")
	if len(lines) < 3 {
		t.Fatal("output should have multiple lines")
	}

	// Check first line has box drawing characters
	firstLine := lines[0]
	if !strings.HasPrefix(firstLine, "┌") || !strings.HasSuffix(firstLine, "┐") {
		t.Error("first line should start with ┌ and end with ┐")
	}

	// Check that the dashboard contains the title line
	hasTitle := false
	for _, line := range lines {
		if strings.Contains(line, "CCQ Dashboard") {
			hasTitle = true
			break
		}
	}
	if !hasTitle {
		t.Error("output should contain title line")
	}
}

func TestGetWindowStatuses(t *testing.T) {
	if !tmux.IsInstalled() {
		t.Skip("tmux not installed")
	}

	tm := tmux.New("ccq-test-statuses")
	if err := tm.NewSession(); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer tm.KillSession()

	q := queue.New(tm)
	dash := dashboard.New(tm, q, 20)

	windows, _ := tm.ListWindows()
	if len(windows) > 0 {
		q.MarkIdle(windows[0].ID)
	}

	// Add another window
	w1, err := tm.NewWindow("/tmp")
	if err != nil {
		t.Fatalf("NewWindow: %v", err)
	}
	q.MarkBusy(w1)

	// Wait for state to propagate
	time.Sleep(100 * time.Millisecond)

	statuses, err := dash.GetWindowStatuses()
	if err != nil {
		t.Fatalf("GetWindowStatuses: %v", err)
	}

	if len(statuses) < 2 {
		t.Fatalf("expected at least 2 windows, got %d", len(statuses))
	}

	// Check that we have different states
	statesSeen := make(map[string]bool)
	for _, s := range statuses {
		statesSeen[s.State] = true
		// Check that directory is set
		if s.Dir == "" {
			t.Error("directory should not be empty")
		}
	}

	// At least check we got some state information
	if len(statesSeen) == 0 {
		t.Error("should have at least some state information")
	}
}
