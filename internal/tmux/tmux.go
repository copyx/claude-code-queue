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
