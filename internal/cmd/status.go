package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jingikim/ccq/internal/queue"
	"github.com/jingikim/ccq/internal/tmux"
)

// Status prints a one-line dashboard summary of all windows.
// Called by tmux status bar via #(ccq _status).
func Status() error {
	tm := tmux.New(sessionName)
	if !tm.HasSession() {
		return nil
	}

	line, err := renderStatusLine(tm)
	if err != nil {
		return err
	}
	fmt.Print(line)
	return nil
}

func renderStatusLine(tm *tmux.Tmux) (string, error) {
	windows, err := tm.ListWindows()
	if err != nil {
		return "", err
	}

	var parts []string
	idleCount := 0

	for _, w := range windows {
		state, _ := tm.GetWindowOption(w.ID, queue.StateKey)

		dir, _ := tm.GetWindowPanePath(w.ID)
		dirName := filepath.Base(dir)
		if dirName == "" || dirName == "." {
			dirName = "~"
		}

		var icon string
		var suffix string

		if w.Active {
			icon = "▶"
		} else {
			switch state {
			case "idle":
				icon = "○"
				idleCount++
				sinceStr, _ := tm.GetWindowOption(w.ID, queue.IdleSinceKey)
				if ts, err := strconv.ParseInt(sinceStr, 10, 64); err == nil && ts > 0 {
					d := time.Since(time.Unix(ts, 0))
					suffix = " " + formatDuration(d)
				}
			case "busy":
				icon = "●"
			default:
				icon = "·"
			}
		}

		parts = append(parts, fmt.Sprintf("%s %s:%s%s", icon, w.Index, dirName, suffix))
	}

	summary := fmt.Sprintf("%d/%d idle", idleCount, len(windows))
	return strings.Join(parts, " | ") + "    " + summary, nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh", int(d.Hours()))
}
