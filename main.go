// main.go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("ccq - Claude Code Queue Manager")
		os.Exit(0)
	}

	switch os.Args[1] {
	case "add":
		fmt.Println("ccq add: not implemented")
	case "_hook":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: ccq _hook <idle|busy|remove>")
			os.Exit(1)
		}
		fmt.Printf("ccq _hook %s: not implemented\n", os.Args[2])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
