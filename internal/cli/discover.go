package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/gallofrancesco1312/hlab-cli/internal/config"
	"github.com/gallofrancesco1312/hlab-cli/pkg/hlabapi"
)

func newDiscoverCommand() *cobra.Command {
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover hlab nodes on the LAN via multicast",
		Long: `Listens for UDP multicast traffic for the specified duration
and saves found nodes to the local cache (~/.hlab/nodes.json).`,

		// RunE is the variant of Run that can return an error.
		// Prefer RunE over Run because cobra handles the error automatically
		// (prints it and sets the exit code).
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiscover(timeout)
		},
	}

	// Local flag: available only on this command, not on subcommands.
	cmd.Flags().DurationVar(&timeout, "timeout", 3*time.Second, "multicast listen duration")

	return cmd
}

func runDiscover(timeout time.Duration) error {
	fmt.Fprintf(os.Stderr, "Listening for multicast beacons for %s...\n", timeout)

	// TODO Phase 4: call the real multicast listener.
	// For now we simulate with a sample node to test the output.
	found := []hlabapi.NodeEntry{
		{
			Node:     "homelab-nas",
			Addr:     "192.168.1.50",
			Port:     8443,
			Version:  "0.1.0",
			LastSeen: time.Now(),
			Stale:    false,
		},
	}

	// Update the on-disk cache before printing.
	existing, err := config.LoadNodes()
	if err != nil {
		return fmt.Errorf("loading node cache: %w", err)
	}
	for _, n := range found {
		existing = config.UpsertNode(existing, n)
	}
	if err := config.SaveNodes(existing); err != nil {
		return fmt.Errorf("saving node cache: %w", err)
	}

	return printNodes(found, globalFlags.output)
}

func newNodesCommand() *cobra.Command {
	var refresh bool

	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "List known nodes from the local cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNodes(refresh)
		},
	}

	cmd.Flags().BoolVar(&refresh, "refresh", false, "force re-discovery before listing")

	return cmd
}

func runNodes(refresh bool) error {
	if refresh {
		if err := runDiscover(3 * time.Second); err != nil {
			return err
		}
	}

	cfg := loadConfig()

	nodes, err := config.LoadNodes()
	if err != nil {
		return fmt.Errorf("loading nodes: %w", err)
	}

	nodes = config.MarkStale(nodes, cfg.StaleAfter)

	return printNodes(nodes, globalFlags.output)
}

// printNodes prints nodes in the requested format.
// Separating output logic from business logic is good practice:
// it makes it easy to add new formats (json, csv) without touching the logic.
func printNodes(nodes []hlabapi.NodeEntry, format outputFormat) error {
	if format == outputJSON {
		return json.NewEncoder(os.Stdout).Encode(nodes)
	}

	if len(nodes) == 0 {
		fmt.Println("No nodes found. Try: hlab discover")
		return nil
	}

	// tabwriter aligns columns automatically.
	// Parameters: output, minwidth, tabwidth, padding, padchar, flags.
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NODE\tADDR\tPORT\tVERSION\tLAST SEEN\tSTALE")
	fmt.Fprintln(w, "----\t----\t----\t-------\t---------\t-----")

	for _, n := range nodes {
		stale := ""
		if n.Stale {
			stale = "yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			n.Node,
			n.Addr,
			n.Port,
			n.Version,
			n.LastSeen.Format(time.RFC3339),
			stale,
		)
	}

	return w.Flush()
}
