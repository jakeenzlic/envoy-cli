package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

// NewGetCmd creates a command to retrieve a value from a vault by key.
func NewGetCmd() *cobra.Command {
	var environment string
	var configPath string

	cmd := &cobra.Command{
		Use:   "get KEY",
		Short: "Get the value of a key from the specified environment vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

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

			value, ok := v.Get(key)
			if !ok {
				return fmt.Errorf("key %q not found in [%s]", key, environment)
			}

			fmt.Fprintln(cmd.OutOrStdout(), value)
			return nil
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "local", "Target environment")
	cmd.Flags().StringVarP(&configPath, "config", "c", ".envoy.json", "Path to config file")
	return cmd
}
