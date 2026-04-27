package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

func NewAuditCmd() *cobra.Command {
	var env string

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Show metadata about vault keys (presence, count, last-modified hint)",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			keys := v.Keys()
			if len(keys) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "vault %q is empty\n", vaultPath)
				return nil
			}

			stat, statErr := os.Stat(vaultPath)
			var modTime string
			if statErr == nil {
				modTime = stat.ModTime().Format(time.RFC3339)
			} else {
				modTime = "unknown"
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "Environment:\t%s\n", env)
			fmt.Fprintf(w, "Vault path:\t%s\n", vaultPath)
			fmt.Fprintf(w, "Key count:\t%d\n", len(keys))
			fmt.Fprintf(w, "Last modified:\t%s\n", modTime)
			fmt.Fprintln(w, "")
			fmt.Fprintln(w, "KEY\tSTATUS")
			for _, k := range keys {
				val, _ := v.Get(k)
				status := "set"
				if val == "" {
					status = "empty"
				}
				fmt.Fprintf(w, "%s\t%s\n", k, status)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVarP(&env, "env", "e", "local", "environment to audit")
	return cmd
}
