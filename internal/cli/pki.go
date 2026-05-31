package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPKICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pki",
		Short: "Gestione certificati mTLS",
	}

	cmd.AddCommand(newPKIInitCommand(), newPKIStatusCommand())
	return cmd
}

func newPKIInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Genera CA, certificato client e certificato server",
		Long: `Crea una CA self-signed e i certificati necessari per mTLS:
  ~/.hlab/ca.crt      — CA pubblica
  ~/.hlab/client.crt  — certificato client
  ~/.hlab/client.key  — chiave privata client`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO Fase 6: implementare generazione PKI reale.
			fmt.Println("pki init: non ancora implementato (Fase 6)")
			return nil
		},
	}
}

func newPKIStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Mostra scadenze dei certificati",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO Fase 6: leggere i cert e mostrare NotBefore/NotAfter.
			fmt.Println("pki status: non ancora implementato (Fase 6)")
			return nil
		},
	}
}
