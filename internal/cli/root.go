// Package cli definisce tutti i comandi della CLI usando cobra.
// Cobra organizza i comandi in un albero: ogni comando può avere sottocomandi,
// flag globali e flag locali.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gallofrancesco1312/hlab-cli/internal/config"
)

// outputFormat è il tipo per il flag --output.
// Usare un tipo custom (invece di string) permette di validare i valori accettati.
type outputFormat string

const (
	outputHuman outputFormat = "human"
	outputJSON  outputFormat = "json"
)

// rootFlags raccoglie i flag globali disponibili su tutti i sottocomandi.
// I flag globali vengono definiti sul root command con PersistentFlags().
type rootFlags struct {
	output outputFormat
}

var globalFlags rootFlags

// NewRootCommand costruisce e ritorna il comando radice.
// In cobra, ogni comando è un *cobra.Command: una struct con Name, Short description,
// Long description, e la funzione Run/RunE da eseguire.
func NewRootCommand(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "hlab",
		Short: "CLI per la gestione remota del tuo homelab",
		Long: `hlab ti permette di scoprire, monitorare e controllare i servizi
del tuo homelab da qualsiasi macchina sulla rete locale.

Esempi:
  hlab discover              # scopre nodi sulla LAN
  hlab nodes                 # lista nodi noti
  hlab nas jellyfin stop     # ferma jellyfin sul nodo "nas"
  hlab pki init              # genera certificati mTLS`,

		// SilenceUsage evita che cobra stampi l'usage completo ad ogni errore.
		// Gli errori vengono già mostrati da cobra; l'usage è rumore.
		SilenceUsage: true,

		// Version fa sì che --version stampi la versione e esca.
		Version: version,
	}

	// PersistentFlags sono flag ereditati da tutti i sottocomandi.
	// StringVarP lega il flag a una variabile: (destinazione, nome, shorthand, default, descrizione)
	root.PersistentFlags().StringVarP(
		(*string)(&globalFlags.output),
		"output", "o",
		string(outputHuman),
		`formato output: "human" (default) o "json"`,
	)

	// Aggiungo i sottocomandi al root. Ogni Add chiama AddCommand internamente.
	root.AddCommand(
		newDiscoverCommand(),
		newNodesCommand(),
		newPKICommand(),
		newConfigCommand(),
	)

	return root
}

// Execute è il punto di ingresso: chiama root.Execute() e gestisce l'uscita.
func Execute(version string) {
	root := NewRootCommand(version)
	if err := root.Execute(); err != nil {
		// cobra stampa già il messaggio di errore; qui usciamo con codice 1.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// loadConfig è un helper condiviso tra i comandi per caricare la config.
// Se il caricamento fallisce, stampa l'errore e termina — comportamento
// corretto per una CLI dove la config è prerequisito.
func loadConfig() config.Config {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "errore config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}
