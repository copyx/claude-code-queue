package cmd

import (
	"github.com/jingikim/ccq/internal/queue"
	"github.com/jingikim/ccq/internal/switcher"
	"github.com/jingikim/ccq/internal/tmux"
)

func Toggle() error {
	tm := tmux.New(sessionName)
	if !tm.HasSession() {
		return nil
	}
	q := queue.New(tm)
	sw := switcher.New(tm, q)

	if sw.IsAutoSwitchOn() {
		sw.SetAutoSwitch(false)
	} else {
		sw.SetAutoSwitch(true)
		// Check queue immediately when enabling
		sw.TrySwitch()
	}
	return nil
}
