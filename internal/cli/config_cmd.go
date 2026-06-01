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
		Short: "Show the current CLI configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadConfig()

			if globalFlags.output == outputJSON {
				return json.NewEncoder(os.Stdout).Encode(cfg)
			}

			// yaml.Marshal produces readable output — handy for debugging.
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("serializing config: %w", err)
			}
			fmt.Print(string(data))
			return nil
		},
	}
}
