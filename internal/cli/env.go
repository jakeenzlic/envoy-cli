package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

// NewEnvCmd returns a command that lists all configured environments
// from the project config, optionally showing vault file paths.
func NewEnvCmd() *cobra.Command {
	var showPaths bool

	cmd := &cobra.Command{
		Use:   "env",
		Short: "List configured environments",
		Long:  `Display all environments defined in the envoy config file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := sync.LoadConfig(".envoy.json")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if len(cfg.Environments) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No environments configured.")
				return nil
			}

			for _, env := range cfg.Environments {
				if showPaths {
					vaultPath := vaultPathForEnv(cfg, env)
					_, statErr := os.Stat(vaultPath)
					status := "ok"
					if os.IsNotExist(statErr) {
						status = "missing"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%-16s %s [%s]\n", env, vaultPath, status)
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), env)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&showPaths, "paths", "p", false, "Show vault file paths and their status")
	return cmd
}

// openVaultForEnv opens (or creates) a vault for the given environment using
// the passphrase resolved from config / environment variable.
func openVaultForEnv(cfg *sync.Config, env string) (*vault.Vault, error) {
	passphrase, err := sync.PassphraseFor(cfg, env)
	if err != nil {
		return nil, fmt.Errorf("passphrase for %q: %w", env, err)
	}
	vaultPath := vaultPathForEnv(cfg, env)
	v, err := vault.New(vaultPath, passphrase)
	if err != nil {
		return nil, fmt.Errorf("open vault for %q: %w", env, err)
	}
	return v, nil
}
