package main

import (
	"fmt"
	"os"

	"github.com/tomo-kay/tene/internal/cli"
)

// These are set via ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.SetVersion(version, commit, date)

	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
