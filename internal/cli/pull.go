package cli

import (
	"fmt"
	"os"

	"github.com/envoy-cli/internal/sync"
	"github.com/envoy-cli/internal/vault"
	"github.com/spf13/cobra"
)

// NewPullCmd returns the pull subcommand which decrypts vault contents
// and writes them to a local .env file.
func NewPullCmd() *cobra.Command {
	var (
		outFile     string
		environment string
		configPath  string
		overwrite   bool
	)

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull environment variables from a vault into a local .env file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := sync.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			passphrase := cfg.Passphrase
			if passphrase == "" {
				return fmt.Errorf("passphrase not set in config or ENVOY_PASSPHRASE")
			}

			if !overwrite {
				if _, err := os.Stat(outFile); err == nil {
					return fmt.Errorf("%s already exists; use --overwrite to replace it", outFile)
				}
			}

			vaultPath := vaultPathForEnv(cfg, environment)
			v, err := vault.New(vaultPath, passphrase)
			if err != nil {
				return fmt.Errorf("opening vault: %w", err)
			}

			syncer := sync.New(v)
			if err := syncer.PullFile(outFile); err != nil {
				return fmt.Errorf("pulling env file: %w", err)
			}

			fmt.Fprintf(os.Stdout, "✓ Pulled %s vault → %s\n", environment, outFile)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outFile, "file", "f", ".env", "Destination .env file path")
	cmd.Flags().StringVarP(&environment, "env", "e", "local", "Source environment (local|staging|production)")
	cmd.Flags().StringVarP(&configPath, "config", "c", "envoy.json", "Path to envoy config file")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite destination file if it exists")

	return cmd
}
