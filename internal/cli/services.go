package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/gallofrancesco1312/hlab-cli/pkg/hlabapi"
)

func newServicesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "services <node>",
		Short: "List services on a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServices(args[0])
		},
	}
}

func runServices(nodeArg string) error {
	cfg := loadConfig()
	node, err := resolveNode(nodeArg, cfg)
	if err != nil {
		return err
	}
	httpClient, err := newHTTPClient(cfg)
	if err != nil {
		return err
	}

	resp, err := httpClient.Get(nodeURL(node, "/services"))
	if err != nil {
		return fmt.Errorf("GET /services: %w", err)
	}
	defer resp.Body.Close()

	var result hlabapi.ServicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	if globalFlags.output == outputJSON {
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	if len(result.Services) == 0 {
		fmt.Printf("No services found on node %s\n", result.Node)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tTYPE\tUPTIME")
	fmt.Fprintln(w, "----\t------\t----\t------")
	for _, svc := range result.Services {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", svc.Name, svc.Status, svc.Type, svc.Uptime)
	}
	return w.Flush()
}

func newStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start <node> <service>",
		Short: "Start a service on a node",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServiceAction(args[0], args[1], "start")
		},
	}
}

func newStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <node> <service>",
		Short: "Stop a service on a node",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServiceAction(args[0], args[1], "stop")
		},
	}
}

func runServiceAction(nodeArg, service, action string) error {
	cfg := loadConfig()
	node, err := resolveNode(nodeArg, cfg)
	if err != nil {
		return err
	}
	httpClient, err := newHTTPClient(cfg)
	if err != nil {
		return err
	}

	url := nodeURL(node, fmt.Sprintf("/services/%s/%s", service, action))
	resp, err := httpClient.Post(url, "application/json", http.NoBody)
	if err != nil {
		return fmt.Errorf("POST %s: %w", url, err)
	}
	defer resp.Body.Close()

	var result hlabapi.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("%s %s: agent returned error", action, service)
	}

	fmt.Printf("%s %s: ok\n", action, service)
	return nil
}
