package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

// NewSetCmd creates a command to set a key-value pair in a vault.
func NewSetCmd() *cobra.Command {
	var environment string
	var configPath string

	cmd := &cobra.Command{
		Use:   "set KEY=VALUE",
		Short: "Set a key-value pair in the specified environment vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parts := strings.SplitN(args[0], "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("argument must be in KEY=VALUE format")
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key == "" {
				return fmt.Errorf("key must not be empty")
			}

			cfg, err := sync.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			passphrase, err := sync.PassphraseFor(environment, cfg)
			if err != nil {
				return fmt.Errorf("passphrase: %w", err)
			}

			vaultPath := vaultPathForEnv(environment, cfg)
			v, err := vault.New(vaultPath, passphrase)
			if err != nil {
				return fmt.Errorf("opening vault: %w", err)
			}

			if err := v.Set(key, value); err != nil {
				return fmt.Errorf("setting key: %w", err)
			}

			if err := v.Save(); err != nil {
				return fmt.Errorf("saving vault: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Set %s in [%s]\n", key, environment)
			return nil
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "local", "Target environment")
	cmd.Flags().StringVarP(&configPath, "config", "c", ".envoy.json", "Path to config file")
	return cmd
}
