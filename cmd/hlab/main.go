// Package main è il punto di ingresso del binario hlab (CLI lato workstation).
//
// In Go ogni eseguibile ha esattamente un package main con una funzione main().
// Il package main non può essere importato da altri package.
package main

import "github.com/gallofrancesco1312/hlab-cli/internal/cli"

// version viene iniettata al momento della compilazione tramite ldflags.
// Il Makefile usa: -ldflags="-X main.version=$(git describe --tags)"
// Questo pattern è standard per versioning senza hard-coding.
var version = "dev"

func main() {
	cli.Execute(version)
}
