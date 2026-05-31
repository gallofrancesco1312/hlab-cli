// Package main è il punto di ingresso del daemon hlab-agent (lato server).
package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	// TODO Fase 2: inizializzare configurazione agente, avviare HTTP server e beacon.
	fmt.Fprintf(os.Stderr, "hlab-agent %s — non ancora implementato (Fase 2)\n", version)
	os.Exit(1)
}
