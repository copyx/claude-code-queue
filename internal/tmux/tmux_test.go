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

	// 세션이 없는 상태 확인
	if tm.HasSession() {
		t.Fatal("session should not exist yet")
	}

	// 세션 생성 (detached)
	if err := tm.NewSession(); err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	defer tm.KillSession()

	// 세션 존재 확인
	if !tm.HasSession() {
		t.Error("session should exist after NewSession")
	}

	// 세션 종료
	if err := tm.KillSession(); err != nil {
		t.Fatalf("KillSession failed: %v", err)
	}
	if tm.HasSession() {
		t.Error("session should not exist after KillSession")
	}
}
