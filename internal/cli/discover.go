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
		Short: "Scopre nodi hlab sulla LAN tramite multicast",
		Long: `Ascolta il traffico multicast UDP per il tempo specificato
e salva i nodi trovati nella cache locale (~/.hlab/nodes.json).`,

		// RunE è la variante di Run che può ritornare un errore.
		// Preferire RunE a Run perché cobra gestisce l'errore automaticamente
		// (lo stampa e imposta il codice di uscita).
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiscover(timeout)
		},
	}

	// Flag locale: disponibile solo su questo comando, non sui sottocomandi.
	cmd.Flags().DurationVar(&timeout, "timeout", 3*time.Second, "durata ascolto multicast")

	return cmd
}

func runDiscover(timeout time.Duration) error {
	fmt.Fprintf(os.Stderr, "Ascolto beacon multicast per %s...\n", timeout)

	// TODO Fase 4: chiamare il listener multicast reale.
	// Per ora simuliamo con un nodo di esempio per testare l'output.
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

	// Aggiorno la cache su disco prima di stampare.
	existing, err := config.LoadNodes()
	if err != nil {
		return fmt.Errorf("caricamento cache nodi: %w", err)
	}
	for _, n := range found {
		existing = config.UpsertNode(existing, n)
	}
	if err := config.SaveNodes(existing); err != nil {
		return fmt.Errorf("salvataggio cache nodi: %w", err)
	}

	return printNodes(found, globalFlags.output)
}

func newNodesCommand() *cobra.Command {
	var refresh bool

	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "Lista i nodi noti nella cache locale",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNodes(refresh)
		},
	}

	cmd.Flags().BoolVar(&refresh, "refresh", false, "forza re-discovery prima di listare")

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
		return fmt.Errorf("caricamento nodi: %w", err)
	}

	nodes = config.MarkStale(nodes, cfg.StaleAfter)

	return printNodes(nodes, globalFlags.output)
}

// printNodes stampa i nodi nel formato richiesto.
// Separare la logica di output dalla logica di business è buona pratica:
// rende facile aggiungere nuovi formati (json, csv) senza toccare la logica.
func printNodes(nodes []hlabapi.NodeEntry, format outputFormat) error {
	if format == outputJSON {
		return json.NewEncoder(os.Stdout).Encode(nodes)
	}

	if len(nodes) == 0 {
		fmt.Println("Nessun nodo trovato. Prova: hlab discover")
		return nil
	}

	// tabwriter allinea le colonne automaticamente.
	// I parametri sono: output, minwidth, tabwidth, padding, padchar, flags.
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
