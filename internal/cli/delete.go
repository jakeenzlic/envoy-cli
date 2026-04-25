package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

// NewDeleteCmd returns a cobra command that deletes a key from a vault.
func NewDeleteCmd() *cobra.Command {
	var env string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a key from the vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			cfg, err := sync.LoadConfig(".envoy.json")
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

			if _, err := v.Get(key); err != nil {
				return fmt.Errorf("key %q not found in vault", key)
			}

			if !yes {
				fmt.Fprintf(cmd.OutOrStdout(), "Delete key %q from %s vault? [y/N]: ", key, env)
				var answer string
				fmt.Fscan(cmd.InOrStdin(), &answer)
				if answer != "y" && answer != "Y" {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil
				}
			}

			if err := v.Delete(key); err != nil {
				return fmt.Errorf("delete key: %w", err)
			}

			if err := v.Save(); err != nil {
				return fmt.Errorf("save vault: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted key %q from %s vault.\n", key, env)
			return nil
		},
	}

	cmd.Flags().StringVarP(&env, "env", "e", "local", "target environment")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompt")
	return cmd
}
