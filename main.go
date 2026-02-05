package main

import (
	"fmt"
	"os"

	"github.com/jingikim/ccq/internal/cmd"
)

func main() {
	var err error

	if len(os.Args) < 2 {
		err = cmd.Root()
	} else {
		switch os.Args[1] {
		case "add":
			err = cmd.Add()
		case "_hook":
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "usage: ccq _hook <idle|busy|remove>")
				os.Exit(1)
			}
			err = cmd.Hook(os.Args[2])
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
			os.Exit(1)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
