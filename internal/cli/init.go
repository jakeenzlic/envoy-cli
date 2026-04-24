package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"envoy-cli/internal/sync"
)

// NewInitCmd creates the `envoy init` command which scaffolds a new envoy
// configuration file and an empty vault for the requested environment.
func NewInitCmd() *cobra.Command {
	var env string
	var vaultDir string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialise envoy in the current project",
		Long: `Creates an envoy.json config file and an empty encrypted vault.
If an envoy.json already exists the command exits without overwriting it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			const configPath = "envoy.json"

			// Refuse to overwrite an existing config.
			if _, err := os.Stat(configPath); err == nil {
				return fmt.Errorf("envoy.json already exists; remove it first if you want to reinitialise")
			}

			if vaultDir == "" {
				vaultDir = ".envoy/vaults"
			}

			// Build a minimal config with a single environment entry.
			cfg := sync.Config{
				Environments: map[string]sync.EnvConfig{
					env: {
						VaultPath:      filepath.Join(vaultDir, env+".vault"),
						PassphraseEnv:  "ENVOY_PASSPHRASE_" + toUpperSnake(env),
					},
				},
			}

			if err := os.MkdirAll(vaultDir, 0o700); err != nil {
				return fmt.Errorf("creating vault directory: %w", err)
			}

			if err := sync.SaveConfig(configPath, cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			fmt.Printf("✔ Created %s\n", configPath)
			fmt.Printf("✔ Vault directory ready: %s\n", vaultDir)
			fmt.Printf("  Set %s to your passphrase before pushing or pulling.\n",
				cfg.Environments[env].PassphraseEnv)
			return nil
		},
	}

	cmd.Flags().StringVarP(&env, "env", "e", "local", "environment name (e.g. local, staging, production)")
	cmd.Flags().StringVar(&vaultDir, "vault-dir", "", "directory where vault files are stored (default .envoy/vaults)")
	return cmd
}

// toUpperSnake converts a string like "staging" to "STAGING".
func toUpperSnake(s string) string {
	out := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		out[i] = c
	}
	return string(out)
}
