package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Mostra la configurazione corrente della CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadConfig()

			if globalFlags.output == outputJSON {
				return json.NewEncoder(os.Stdout).Encode(cfg)
			}

			// yaml.Marshal produce output YAML leggibile — comodo per debug.
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("serializzazione config: %w", err)
			}
			fmt.Print(string(data))
			return nil
		},
	}
}
