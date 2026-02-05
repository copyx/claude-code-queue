// Package tmux wraps tmux commands for managing named sessions.
package tmux

import (
	"os/exec"
	"strings"
)

// IsInstalled checks if tmux is available in PATH.
func IsInstalled() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// Tmux wraps tmux commands for a named session.
type Tmux struct {
	Session string
}

// New creates a Tmux instance for the given session name.
func New(session string) *Tmux {
	return &Tmux{Session: session}
}

func (t *Tmux) run(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Run executes an arbitrary tmux command.
func (t *Tmux) Run(args ...string) (string, error) {
	return t.run(args...)
}

// HasSession returns true if the named session exists.
func (t *Tmux) HasSession() bool {
	_, err := t.run("has-session", "-t", t.Session)
	return err == nil
}

// NewSession creates a new detached session.
func (t *Tmux) NewSession() error {
	_, err := t.run("new-session", "-d", "-s", t.Session)
	return err
}

// KillSession destroys the session.
func (t *Tmux) KillSession() error {
	_, err := t.run("kill-session", "-t", t.Session)
	return err
}

// NewWindow creates a new window in the session running the default shell
// in the given directory. Returns the window ID.
func (t *Tmux) NewWindow(dir string) (string, error) {
	out, err := t.run("new-window", "-d", "-t", t.Session, "-c", dir, "-P", "-F", "#{window_id}")
	if err != nil {
		return "", err
	}
	return out, nil
}

// WindowInfo holds metadata about a tmux window.
type WindowInfo struct {
	ID     string
	Index  string
	Name   string
	Active bool
}

// ListWindows returns all windows in the session.
func (t *Tmux) ListWindows() ([]WindowInfo, error) {
	out, err := t.run("list-windows", "-t", t.Session, "-F", "#{window_id}\t#{window_index}\t#{window_name}\t#{window_active}")
	if err != nil {
		return nil, err
	}
	var windows []WindowInfo
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 4)
		if len(parts) < 4 {
			continue
		}
		windows = append(windows, WindowInfo{
			ID:     parts[0],
			Index:  parts[1],
			Name:   parts[2],
			Active: parts[3] == "1",
		})
	}
	return windows, nil
}

// SetWindowOption sets a user option on a window.
func (t *Tmux) SetWindowOption(windowID, key, value string) error {
	_, err := t.run("set-option", "-w", "-t", windowID, key, value)
	return err
}

// GetWindowOption reads a user option from a window. Returns "" if not set.
func (t *Tmux) GetWindowOption(windowID, key string) (string, error) {
	out, err := t.run("show-options", "-w", "-v", "-t", windowID, key)
	if err != nil {
		return "", nil // option not set
	}
	return out, nil
}

// SelectWindow switches the active window.
func (t *Tmux) SelectWindow(windowID string) error {
	_, err := t.run("select-window", "-t", windowID)
	return err
}

// ActiveWindowID returns the window ID of the currently active window.
func (t *Tmux) ActiveWindowID() (string, error) {
	out, err := t.run("display-message", "-t", t.Session, "-p", "#{window_id}")
	if err != nil {
		return "", err
	}
	return out, nil
}

// SetSessionOption sets a session-level option.
func (t *Tmux) SetSessionOption(key, value string) error {
	_, err := t.run("set-option", "-t", t.Session, key, value)
	return err
}

// GetSessionOption reads a session-level option.
func (t *Tmux) GetSessionOption(key string) (string, error) {
	out, err := t.run("show-options", "-v", "-t", t.Session, key)
	if err != nil {
		return "", nil
	}
	return out, nil
}

// SendKeys sends keystrokes to a window. If enter is true, appends Enter.
func (t *Tmux) SendKeys(target, keys string, enter bool) error {
	args := []string{"send-keys", "-t", target, keys}
	if enter {
		args = append(args, "Enter")
	}
	_, err := t.run(args...)
	return err
}

// WindowIDFromPane returns the window ID containing the given pane.
func (t *Tmux) WindowIDFromPane(paneID string) (string, error) {
	out, err := t.run("display-message", "-t", paneID, "-p", "#{window_id}")
	if err != nil {
		return "", err
	}
	return out, nil
}
