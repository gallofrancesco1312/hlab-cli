// Package main is the entry point for the hlab-agent daemon (server side).
package main

import (
	"os"
	"log/slog"

	"github.com/gallofrancesco1312/hlab-cli/internal/agent"
)

func main() {
	slog.Info("Starting hlab-agent...")
	agent.NewServer(os.Args[1]).Start()
}
