package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jingikim/ccq/internal/config"
	"github.com/jingikim/ccq/internal/tmux"
)

const sessionName = "ccq"

func Root() error {
	if !tmux.IsInstalled() {
		return fmt.Errorf("tmux is not installed. Install it with: brew install tmux")
	}

	tm := tmux.New(sessionName)

	// 이미 세션 존재하면 attach/switch
	if tm.HasSession() {
		return attachOrSwitch(tm)
	}

	// 설정 로드
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 최초 실행: prefix 설정
	if cfg.Prefix == "" {
		cfg.Prefix = promptPrefix()
		if err := config.Save(config.DefaultPath(), cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	// tmux 세션 생성
	if err := tm.NewSession(); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// tmux 설정 적용
	tm.SetSessionOption("@ccq_auto_switch", "on")
	tm.SetSessionOption("remain-on-exit", "off")
	tm.SetSessionOption("prefix", cfg.Prefix)

	// 상태바 설정
	tm.SetSessionOption("status-left", "[#{?@ccq_auto_switch,AUTO,MANUAL}] ")
	tm.SetSessionOption("status-right", "#{session_windows} windows")
	tm.SetSessionOption("status-style", "bg=colour236,fg=colour248")
	tm.SetSessionOption("window-status-current-format", "#[fg=colour214,bold]#W")
	tm.SetSessionOption("window-status-format", "#W")

	// prefix + a 로 자동 전환 토글
	tm.Run("bind-key", "-T", "prefix", "a", "run-shell", "ccq _toggle")

	// 첫 window에서 claude 실행
	windows, _ := tm.ListWindows()
	if len(windows) > 0 {
		tm.SendKeys(windows[0].ID, "claude", true)
	}

	return attachOrSwitch(tm)
}

func attachOrSwitch(tm *tmux.Tmux) error {
	if os.Getenv("TMUX") != "" {
		cmd := exec.Command("tmux", "switch-client", "-t", sessionName)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	cmd := exec.Command("tmux", "attach-session", "-t", sessionName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func promptPrefix() string {
	fmt.Println("ccq tmux prefix 키를 선택하세요 (Claude Code Ctrl+B 충돌 방지):")
	fmt.Println("  1) Ctrl+Space (권장)")
	fmt.Println("  2) Ctrl+\\")
	fmt.Println("  3) Ctrl+A")
	fmt.Print("  선택 [1]: ")

	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "2":
		return "C-\\"
	case "3":
		return "C-a"
	default:
		return "C-Space"
	}
}
