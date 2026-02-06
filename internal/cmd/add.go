package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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

	// Save current active window before creating new one
	activeID, _ := tm.ActiveWindowID()

	windowID, err := tm.NewWindow(dir)
	if err != nil {
		return fmt.Errorf("failed to create window: %w", err)
	}

	if err := tm.SendKeys(windowID, "claude", true); err != nil {
		return fmt.Errorf("failed to start claude: %w", err)
	}

	// Mark return target so HandleIdle can switch back after initial setup
	inTmux := os.Getenv("TMUX") != ""
	if inTmux {
		tm.SetWindowOption(windowID, "@ccq_return_to", activeID)
	} else {
		tty := getTTY()
		if tty != "" {
			tm.SetWindowOption(windowID, "@ccq_return_to", "__detach__:"+tty)
		} else {
			tm.SetWindowOption(windowID, "@ccq_return_to", "__detach__")
		}
	}

	// Switch to new window so user can handle initial setup (trust prompt, etc.)
	tm.SelectWindow(windowID)

	fmt.Printf("세션 추가됨: %s (%s)\n", windowID, dir)

	if !inTmux {
		// Attach to session (blocks until hook detaches this client)
		cmd := exec.Command("tmux", "attach-session", "-t", sessionName)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return nil
}

func getTTY() string {
	cmd := exec.Command("tty")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
