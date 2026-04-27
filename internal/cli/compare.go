package cli

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

func NewCompareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compare <env1> <env2>",
		Short: "Compare keys present in two environments",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			env1, env2 := args[0], args[1]

			cfg, err := sync.LoadConfig(".envoy.json")
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			v1, err := openVault(env1, cfg)
			if err != nil {
				return fmt.Errorf("open %s vault: %w", env1, err)
			}
			v2, err := openVault(env2, cfg)
			if err != nil {
				return fmt.Errorf("open %s vault: %w", env2, err)
			}

			keys1 := toSet(v1.Keys())
			keys2 := toSet(v2.Keys())
			all := mergeKeys(keys1, keys2)

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "KEY\t%s\t%s\n", env1, env2)
			for _, k := range all {
				c1, c2 := presence(keys1, k), presence(keys2, k)
				if c1 != c2 {
					fmt.Fprintf(w, "%s\t%s\t%s\n", k, c1, c2)
				}
			}
			return w.Flush()
		},
	}
	return cmd
}

func openVault(env string, cfg *sync.Config) (*vault.Vault, error) {
	passphrase, err := sync.PassphraseFor(env, cfg)
	if err != nil {
		return nil, err
	}
	return vault.New(vaultPathForEnv(env, cfg), passphrase)
}

func toSet(keys []string) map[string]struct{} {
	m := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		m[k] = struct{}{}
	}
	return m
}

func mergeKeys(a, b map[string]struct{}) []string {
	seen := make(map[string]struct{})
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func presence(m map[string]struct{}, k string) string {
	if _, ok := m[k]; ok {
		return "✓"
	}
	return "✗"
}
