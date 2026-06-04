// Package cli defines all CLI commands using cobra.
// Cobra organizes commands in a tree: each command can have subcommands,
// global flags, and local flags.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gallofrancesco1312/hlab-cli/internal/config"
)

// outputFormat is the type for the --output flag.
// Using a custom type (instead of string) allows validating accepted values.
type outputFormat string

const (
	outputHuman outputFormat = "human"
	outputJSON  outputFormat = "json"
)

// rootFlags holds the global flags available on all subcommands.
// Global flags are defined on the root command with PersistentFlags().
type rootFlags struct {
	output outputFormat
}

var globalFlags rootFlags

// NewRootCommand builds and returns the root command.
// In cobra, every command is a *cobra.Command: a struct with Name, Short description,
// Long description, and the Run/RunE function to execute.
func NewRootCommand(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "hlab",
		Short: "CLI for remote management of your homelab",
		Long: `hlab lets you discover, monitor, and control services
in your homelab from any machine on the local network.

Examples:
  hlab discover              # discover nodes on the LAN
  hlab nodes                 # list known nodes
  hlab services nas          # list services on node "nas"
  hlab start nas jellyfin    # start jellyfin on node "nas"
  hlab stop nas jellyfin     # stop jellyfin on node "nas"
  hlab pki init              # generate mTLS certificates`,

		// SilenceUsage prevents cobra from printing the full usage on every error.
		// Errors are already shown by cobra; the usage is just noise.
		SilenceUsage: true,

		// Version makes --version print the version and exit.
		Version: version,
	}

	// PersistentFlags are flags inherited by all subcommands.
	// StringVarP binds the flag to a variable: (destination, name, shorthand, default, description)
	root.PersistentFlags().StringVarP(
		(*string)(&globalFlags.output),
		"output", "o",
		string(outputHuman),
		`output format: "human" (default) or "json"`,
	)

	root.AddCommand(
		newDiscoverCommand(),
		newNodesCommand(),
		newServicesCommand(),
		newStartCommand(),
		newStopCommand(),
		newPKICommand(),
		newConfigCommand(),
	)

	return root
}

// Execute is the entry point: calls root.Execute() and handles exit.
func Execute(version string) {
	root := NewRootCommand(version)
	if err := root.Execute(); err != nil {
		// cobra already prints the error message; here we exit with code 1.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// loadConfig is a shared helper for commands to load the config.
// If loading fails, it prints the error and terminates — correct behavior
// for a CLI where config is a prerequisite.
func loadConfig() config.Config {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}
	return cfg
}
