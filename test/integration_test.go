// test/integration_test.go
package test

import (
	"testing"
	"time"

	"github.com/jingikim/ccq/internal/queue"
	"github.com/jingikim/ccq/internal/switcher"
	"github.com/jingikim/ccq/internal/tmux"
)

func TestFullFlow(t *testing.T) {
	if !tmux.IsInstalled() {
		t.Skip("tmux not installed")
	}

	name := "ccq-integration-test"
	tm := tmux.New(name)

	// Start from a clean state
	tm.KillSession()

	if err := tm.NewSession(); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer tm.KillSession()

	q := queue.New(tm)
	sw := switcher.New(tm, q)
	sw.SetAutoSwitch(true)

	// 3 windows
	windows, _ := tm.ListWindows()
	w0 := windows[0].ID
	w1, _ := tm.NewWindow("/tmp")
	w2, _ := tm.NewWindow("/tmp")

	// Scenario: w1=idle(first), w0=idle(later), w2=busy
	q.MarkIdle(w1)
	time.Sleep(time.Second)
	q.MarkIdle(w0)
	q.MarkBusy(w2)

	// w0 is active and becomes busy -> switch to w1 (oldest idle)
	q.MarkBusy(w0)
	sw.TrySwitch()

	activeID, _ := tm.ActiveWindowID()
	if activeID != w1 {
		t.Errorf("step1: expected switch to w1 (%s), got %s", w1, activeID)
	}

	// w1 also becomes busy -> no idle windows, no switch
	q.MarkBusy(w1)
	switched := sw.TrySwitch()
	if switched {
		t.Error("step2: expected no switch when no idle windows")
	}

	// w2 becomes idle -> current w1 is busy -> switch to w2
	q.MarkIdle(w2)
	sw.TrySwitch()

	activeID, _ = tm.ActiveWindowID()
	if activeID != w2 {
		t.Errorf("step3: expected switch to w2 (%s), got %s", w2, activeID)
	}

	// auto-switch OFF -> no switch
	sw.SetAutoSwitch(false)
	q.MarkIdle(w0)
	q.MarkBusy(w2)
	switched = sw.TrySwitch()
	if switched {
		t.Error("step4: expected no switch when auto-switch is off")
	}
}
