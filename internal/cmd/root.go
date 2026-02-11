package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jingikim/ccq/internal/config"
	"github.com/jingikim/ccq/internal/tmux"
)

const (
	sessionName   = "ccq"
	configVersion = "3" // Increment when session settings change (keybindings, status bar, etc.)
)

// initSessionSettings applies all settings for a newly created session.
func initSessionSettings(tm *tmux.Tmux, prefix string) error {
	tm.SetSessionOption("@ccq_auto_switch", "on")
	tm.SetSessionOption("remain-on-exit", "off")
	if err := tm.SetSessionOption("prefix", prefix); err != nil {
		return fmt.Errorf("failed to set prefix key %q: %w", prefix, err)
	}
	applyVersionedSettings(tm)
	tm.SetSessionOption("@ccq_config_version", configVersion)
	return nil
}

// migrateSessionSettings updates only versioned settings without touching user preferences.
// Preserves: prefix, @ccq_auto_switch, remain-on-exit
func migrateSessionSettings(tm *tmux.Tmux) {
	applyVersionedSettings(tm)
	tm.SetSessionOption("@ccq_config_version", configVersion)
}

// applyVersionedSettings applies settings that may change between versions.
// These are safe to re-apply without affecting user state.
func applyVersionedSettings(tm *tmux.Tmux) {
	// Status bar (line 0 - bottom)
	tm.SetSessionOption("status-left", "[#{?#{==:#{@ccq_auto_switch},on},AUTO,MANUAL}] ")
	tm.SetSessionOption("status-right", "#{session_windows} windows")
	tm.SetSessionOption("status-style", "bg=colour236,fg=colour248")
	tm.SetSessionOption("window-status-current-format", "#[fg=colour214,bold]#I:#{b:pane_current_path}#{?#{@ccq_state}, #{@ccq_state},}")
	tm.SetSessionOption("window-status-format", "#I:#{b:pane_current_path}#{?#{@ccq_state}, #{@ccq_state},}")

	// Dashboard status bar (line 1 - top)
	tm.SetSessionOption("status", "2")
	tm.SetSessionOption("status-interval", "2")
	tm.SetSessionOption("status-format[1]", "#[align=left]#(ccq _status)")

	// Keybindings
	tm.Run("bind-key", "-T", "prefix", "a", "run-shell", "ccq _toggle")
	tm.Run("bind-key", "-T", "prefix", "g", "run-shell", "ccq toggle-dashboard")
}

func Root() error {
	if !tmux.IsInstalled() {
		return fmt.Errorf("tmux is not installed. Install it with: brew install tmux")
	}

	tm := tmux.New(sessionName)

	// 이미 세션 존재하면 윈도우 추가
	if tm.HasSession() {
		return addWindow(tm)
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

	// Apply all session settings
	if err := initSessionSettings(tm, cfg.Prefix); err != nil {
		tm.KillSession()
		return err
	}

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

func addWindow(tm *tmux.Tmux) error {
	// Check if session configuration needs migration
	currentVersion, _ := tm.GetSessionOption("@ccq_config_version")
	if currentVersion != configVersion {
		migrateSessionSettings(tm)
		fmt.Printf("✓ ccq settings updated (v%s → v%s)\n", currentVersion, configVersion)
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

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

	tm.SelectWindow(windowID)

	if !inTmux {
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
