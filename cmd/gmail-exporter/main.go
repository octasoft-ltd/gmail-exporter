package main

import (
	"fmt"
	"os"

	"github.com/octasoft-ltd/gmail-exporter/internal/cli"
)

// Version information (set by build flags)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Set version information for CLI
	cli.SetVersion(version, commit, date)

	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
