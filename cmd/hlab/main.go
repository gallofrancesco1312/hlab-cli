// Package main is the entry point for the hlab binary (workstation-side CLI).
//
// In Go every executable has exactly one package main with a main() function.
// Package main cannot be imported by other packages.
package main

import "github.com/gallofrancesco1312/hlab-cli/internal/cli"

// version is injected at build time via ldflags.
// The Makefile uses: -ldflags="-X main.version=$(git describe --tags)"
// This pattern is standard for versioning without hard-coding.
var version = "dev"

func main() {
	cli.Execute(version)
}
