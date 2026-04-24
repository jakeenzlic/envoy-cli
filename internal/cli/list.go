package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

// NewListCmd returns a cobra command that lists all keys stored in a vault
// for the given environment.
func NewListCmd() *cobra.Command {
	var showValues bool

	cmd := &cobra.Command{
		Use:   "list [environment]",
		Short: "List all keys stored in a vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			env := args[0]

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

			keys := v.Keys()
			if len(keys) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No keys found in %q vault.\n", env)
				return nil
			}

			sort.Strings(keys)

			for _, k := range keys {
				if showValues {
					val, _ := v.Get(k)
					fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", k, val)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", k)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&showValues, "values", "v", false, "also print the values (use with care)")
	return cmd
}
