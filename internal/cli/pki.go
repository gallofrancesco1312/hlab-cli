package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPKICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pki",
		Short: "Manage mTLS certificates",
	}

	cmd.AddCommand(newPKIInitCommand(), newPKIStatusCommand())
	return cmd
}

func newPKIInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Generate CA, client certificate, and server certificate",
		Long: `Creates a self-signed CA and the certificates required for mTLS:
  ~/.hlab/ca.crt      — public CA
  ~/.hlab/client.crt  — client certificate
  ~/.hlab/client.key  — client private key`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO Phase 6: implement real PKI generation.
			fmt.Println("pki init: not yet implemented (Phase 6)")
			return nil
		},
	}
}

func newPKIStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show certificate expiry dates",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO Phase 6: read certs and show NotBefore/NotAfter.
			fmt.Println("pki status: not yet implemented (Phase 6)")
			return nil
		},
	}
}
