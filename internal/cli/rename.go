package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

// NewRenameCmd creates the rename command which renames a key within a vault.
func NewRenameCmd() *cobra.Command {
	var env string

	cmd := &cobra.Command{
		Use:   "rename <old-key> <new-key>",
		Short: "Rename a key in the vault",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldKey := args[0]
			newKey := args[1]

			if oldKey == newKey {
				return fmt.Errorf("old and new key names are identical: %q", oldKey)
			}

			cfg, err := sync.LoadConfig("envoy.json")
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			passphrase, err := sync.PassphraseFor(env, cfg)
			if err != nil {
				return fmt.Errorf("passphrase: %w", err)
			}

			vaultPath := vaultPathForEnv(env, cfg)
			v, err := vault.New(vaultPath, passphrase)
			if err != nil {
				return fmt.Errorf("open vault: %w", err)
			}

			val, ok := v.Get(oldKey)
			if !ok {
				return fmt.Errorf("key %q not found in vault", oldKey)
			}

			if _, exists := v.Get(newKey); exists {
				return fmt.Errorf("key %q already exists; delete it first or use --overwrite", newKey)
			}

			v.Set(newKey, val)
			v.Delete(oldKey)

			if err := v.Save(); err != nil {
				return fmt.Errorf("save vault: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Renamed %q → %q in %s vault\n", oldKey, newKey, env)
			return nil
		},
	}

	cmd.Flags().StringVarP(&env, "env", "e", "local", "target environment")
	return cmd
}
