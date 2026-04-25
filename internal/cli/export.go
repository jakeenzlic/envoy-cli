package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

// NewExportCmd returns a cobra command that exports vault secrets to a .env file.
func NewExportCmd() *cobra.Command {
	var (
		env      string
		outFile  string
		quoted   bool
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export vault secrets to a .env file",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			sort.Strings(keys)

			var sb strings.Builder
			for _, k := range keys {
				val, _ := v.Get(k)
				if quoted {
					fmt.Fprintf(&sb, "%s=%q\n", k, val)
				} else {
					fmt.Fprintf(&sb, "%s=%s\n", k, val)
				}
			}

			if outFile == "" || outFile == "-" {
				fmt.Fprint(cmd.OutOrStdout(), sb.String())
				return nil
			}

			if err := os.WriteFile(outFile, []byte(sb.String()), 0600); err != nil {
				return fmt.Errorf("write file: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Exported %d key(s) to %s\n", len(keys), outFile)
			return nil
		},
	}

	cmd.Flags().StringVarP(&env, "env", "e", "local", "Environment to export from")
	cmd.Flags().StringVarP(&outFile, "output", "o", "-", "Output file path (default: stdout)")
	cmd.Flags().BoolVar(&quoted, "quoted", false, "Quote values using Go string quoting")

	return cmd
}
