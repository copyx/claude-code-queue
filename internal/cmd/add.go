package cmd

import (
	"fmt"
	"os"

	"github.com/jingikim/ccq/internal/tmux"
)

func Add() error {
	tm := tmux.New(sessionName)

	if !tm.HasSession() {
		return fmt.Errorf("ccq 세션이 없습니다. 먼저 ccq를 실행하세요.")
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	windowID, err := tm.NewWindow(dir)
	if err != nil {
		return fmt.Errorf("failed to create window: %w", err)
	}

	if err := tm.SendKeys(windowID, "claude", true); err != nil {
		return fmt.Errorf("failed to start claude: %w", err)
	}

	fmt.Printf("세션 추가됨: %s (%s)\n", windowID, dir)
	return nil
}
