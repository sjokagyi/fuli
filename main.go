package main

import (
	"fmt"
	"os"

	"github.com/sjokagyi/fuli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		// If the command fails (and hasn't handled its own output), print error and exit.
		// Note: The TUI usually handles its own errors visually, this is a fallback.
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
