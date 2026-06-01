// Package main is the entry point for the hlab-agent daemon (server side).
package main

import (
	"os"

	"github.com/gallofrancesco1312/hlab-cli/internal/agent"
)

func main() {
	agent.NewServer(os.Args[1]).Start()
}
