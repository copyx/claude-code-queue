// Package dashboard provides real-time display of window states.
package dashboard

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jingikim/ccq/internal/queue"
	"github.com/jingikim/ccq/internal/tmux"
)

const (
	stateKey     = "@ccq_state"
	idleSinceKey = "@ccq_idle_since"
)

// Dashboard displays window state information.
type Dashboard struct {
	tm    *tmux.Tmux
	q     *queue.Queue
	width int
}

// New creates a Dashboard for the given tmux session and queue.
func New(tm *tmux.Tmux, q *queue.Queue, width int) *Dashboard {
	if width <= 0 {
		width = 20
	}
	return &Dashboard{tm: tm, q: q, width: width}
}

// WindowStatus represents the state of a single window.
type WindowStatus struct {
	Index     string
	State     string // "busy", "idle", or "active"
	IdleSince time.Time
	Dir       string
}

// GetWindowStatuses returns status information for all windows.
func (d *Dashboard) GetWindowStatuses() ([]WindowStatus, error) {
	windows, err := d.tm.ListWindows()
	if err != nil {
		return nil, err
	}

	var statuses []WindowStatus
	for _, w := range windows {
		state, _ := d.tm.GetWindowOption(w.ID, stateKey)
		if state == "" {
			state = "unknown"
		}

		var idleSince time.Time
		if state == "idle" {
			sinceStr, _ := d.tm.GetWindowOption(w.ID, idleSinceKey)
			if ts, err := strconv.ParseInt(sinceStr, 10, 64); err == nil && ts > 0 {
				idleSince = time.Unix(ts, 0)
			}
		}

		dir, _ := d.tm.GetWindowPanePath(w.ID)
		dirName := filepath.Base(dir)
		if dirName == "" || dirName == "." {
			dirName = "~"
		}

		status := WindowStatus{
			Index:     w.Index,
			State:     state,
			IdleSince: idleSince,
			Dir:       dirName,
		}

		// Mark active window
		if w.Active {
			status.State = "active"
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

// Render formats the dashboard output.
func (d *Dashboard) Render() (string, error) {
	statuses, err := d.GetWindowStatuses()
	if err != nil {
		return "", err
	}

	var lines []string
	lines = append(lines, d.line("┌", "─", "┐"))
	lines = append(lines, d.centeredLine("CCQ Dashboard"))
	lines = append(lines, d.line("├", "─", "┤"))

	for _, s := range statuses {
		line := d.formatWindowLine(s)
		lines = append(lines, line)
	}

	lines = append(lines, d.line("└", "─", "┘"))

	return strings.Join(lines, "\n"), nil
}

// formatWindowLine formats a single window status line.
func (d *Dashboard) formatWindowLine(s WindowStatus) string {
	var icon string
	var idleInfo string

	switch s.State {
	case "busy":
		icon = "●"
	case "idle":
		icon = "○"
		if !s.IdleSince.IsZero() {
			duration := time.Since(s.IdleSince)
			idleInfo = formatDuration(duration)
		}
	case "active":
		icon = "▶"
	default:
		icon = "?"
	}

	// Format: │ W0: dir    ● 3m │
	content := fmt.Sprintf("W%s: %s", s.Index, s.Dir)
	if idleInfo != "" {
		content = fmt.Sprintf("%s %s", content, idleInfo)
	}

	// Pad to width, reserving space for icon
	maxContentWidth := d.width - 6 // 2 for "│ ", 2 for " ", 1 for icon, 1 for "│"
	if maxContentWidth < 1 {
		maxContentWidth = 1 // Ensure minimum width
	}
	if len(content) > maxContentWidth {
		// Truncate to fit, accounting for ellipsis
		if maxContentWidth > 1 {
			content = content[:maxContentWidth-1] + "…"
		} else {
			content = "…"
		}
	}

	padding := maxContentWidth - len(content)
	if padding < 0 {
		padding = 0 // Prevent negative padding
	}
	return fmt.Sprintf("│ %s%s %s │", content, strings.Repeat(" ", padding), icon)
}

// line creates a horizontal border line.
func (d *Dashboard) line(left, mid, right string) string {
	return fmt.Sprintf("%s%s%s", left, strings.Repeat(mid, d.width-2), right)
}

// centeredLine creates a centered text line with borders.
func (d *Dashboard) centeredLine(text string) string {
	available := d.width - 4 // 2 for "│ " and " │"
	if len(text) > available {
		text = text[:available]
	}
	padding := available - len(text)
	leftPad := padding / 2
	rightPad := padding - leftPad
	return fmt.Sprintf("│ %s%s%s │", strings.Repeat(" ", leftPad), text, strings.Repeat(" ", rightPad))
}

// formatDuration formats a duration as "3m", "5s", etc.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh", int(d.Hours()))
}
