package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

// NewSearchCmd returns a cobra command that searches for keys (and optionally values)
// in the vault for a given environment.
func NewSearchCmd() *cobra.Command {
	var searchValues bool
	var caseSensitive bool

	cmd := &cobra.Command{
		Use:   "search <env> <pattern>",
		Short: "Search for keys (or values) matching a pattern in a vault",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			env := args[0]
			pattern := args[1]

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

			keys := v.Keys().All()
			if len(keys) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "vault is empty")
				return nil
			}

			matchFn := func(s string) bool {
				if caseSensitive {
					return strings.Contains(s, pattern)
				}
				return strings.Contains(strings.ToLower(s), strings.ToLower(pattern))
			}

			found := 0
			for _, key := range keys {
				if matchFn(key) {
					if searchValues {
						val, _ := v.Get(key)
						fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", key, val)
					} else {
						fmt.Fprintln(cmd.OutOrStdout(), key)
					}
					found++
					continue
				}
				if searchValues {
					val, _ := v.Get(key)
					if matchFn(val) {
						fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", key, val)
						found++
					}
				}
			}

			if found == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "no matches for %q\n", pattern)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&searchValues, "values", "v", false, "also search inside values")
	cmd.Flags().BoolVarP(&caseSensitive, "case-sensitive", "c", false, "use case-sensitive matching")
	return cmd
}
