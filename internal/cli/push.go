package cli

import (
	"fmt"
	"os"

	"github.com/envoy-cli/internal/sync"
	"github.com/envoy-cli/internal/vault"
	"github.com/spf13/cobra"
)

// NewPushCmd returns the push subcommand which encrypts and uploads
// local .env variables into the target environment vault.
func NewPushCmd() *cobra.Command {
	var (
		envFile     string
		environment string
		configPath  string
	)

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push local .env file to a remote environment vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := sync.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			passphrase := cfg.Passphrase
			if passphrase == "" {
				return fmt.Errorf("passphrase not set in config or ENVOY_PASSPHRASE")
			}

			vaultPath := vaultPathForEnv(cfg, environment)
			v, err := vault.New(vaultPath, passphrase)
			if err != nil {
				return fmt.Errorf("opening vault: %w", err)
			}

			syncer := sync.New(v)
			if err := syncer.PushFile(envFile); err != nil {
				return fmt.Errorf("pushing env file: %w", err)
			}

			fmt.Fprintf(os.Stdout, "✓ Pushed %s → %s vault\n", envFile, environment)
			return nil
		},
	}

	cmd.Flags().StringVarP(&envFile, "file", "f", ".env", "Path to the .env file")
	cmd.Flags().StringVarP(&environment, "env", "e", "local", "Target environment (local|staging|production)")
	cmd.Flags().StringVarP(&configPath, "config", "c", "envoy.json", "Path to envoy config file")

	return cmd
}

func vaultPathForEnv(cfg *sync.Config, environment string) string {
	if path, ok := cfg.Environments[environment]; ok {
		return path
	}
	return environment + ".vault"
}
