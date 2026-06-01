// Package main is the entry point for the hlab-agent daemon (server side).
package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	// TODO Phase 2: initialize agent configuration, start HTTP server and beacon.
	fmt.Fprintf(os.Stderr, "hlab-agent %s — not yet implemented (Phase 2)\n", version)
	os.Exit(1)
}
