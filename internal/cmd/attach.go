package cmd

import (
	"fmt"
	"os"

	"github.com/jingikim/ccq/internal/tmux"
)

// Attach connects to an existing ccq session without creating a new window.
func Attach() error {
	tm := tmux.New(sessionName)
	if !tm.HasSession() {
		fmt.Fprintln(os.Stderr, "ccq: no active session")
		os.Exit(1)
	}
	return attachOrSwitch(tm)
}
