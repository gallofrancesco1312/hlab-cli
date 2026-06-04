package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/gallofrancesco1312/hlab-cli/internal/config"
	"github.com/gallofrancesco1312/hlab-cli/pkg/hlabapi"
)

func newDiscoverCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Probe configured nodes and update the local cache",
		Long: `Contacts each node listed under 'nodes:' in ~/.hlab/config.yaml,
calls GET /health, and saves results to ~/.hlab/nodes.json.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiscover()
		},
	}
	return cmd
}

func runDiscover() error {
	cfg := loadConfig()

	if len(cfg.Nodes) == 0 {
		return fmt.Errorf("no nodes configured — add entries under 'nodes:' in ~/.hlab/config.yaml")
	}

	client, err := newHTTPClient(cfg)
	if err != nil {
		return fmt.Errorf("building HTTP client: %w", err)
	}

	found, err := probeNodes(cfg.Nodes, client)
	if err != nil {
		return err
	}

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

func probeNodes(refs []config.NodeRef, client *http.Client) ([]hlabapi.NodeEntry, error) {
	var entries []hlabapi.NodeEntry
	for _, ref := range refs {
		url := fmt.Sprintf("https://%s:%d/health", ref.Addr, ref.Port)
		resp, err := client.Get(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: %s:%d unreachable: %v\n", ref.Addr, ref.Port, err)
			continue
		}

		var h hlabapi.HealthResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&h)
		resp.Body.Close()
		if decodeErr != nil {
			fmt.Fprintf(os.Stderr, "warn: %s:%d bad response: %v\n", ref.Addr, ref.Port, decodeErr)
			continue
		}

		entries = append(entries, hlabapi.NodeEntry{
			Node:     h.Node,
			Addr:     ref.Addr,
			Port:     ref.Port,
			Version:  h.Version,
			LastSeen: time.Now(),
		})
	}
	return entries, nil
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
		if err := runDiscover(); err != nil {
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

func printNodes(nodes []hlabapi.NodeEntry, format outputFormat) error {
	if format == outputJSON {
		return json.NewEncoder(os.Stdout).Encode(nodes)
	}

	if len(nodes) == 0 {
		fmt.Println("No nodes found. Try: hlab discover")
		return nil
	}

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
