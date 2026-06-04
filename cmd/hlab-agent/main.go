// Package main is the entry point for the hlab-agent daemon (server side).
package main

import (
	"os"
	"log/slog"

	"github.com/gallofrancesco1312/hlab-cli/internal/agent"
)

func main() {
	slog.Info("Starting hlab-agent...")
	if err := agent.NewServer(os.Args[1]).Start(); err != nil {
		slog.Error("Agent stopped", "error", err)
		os.Exit(1)
	}
}
