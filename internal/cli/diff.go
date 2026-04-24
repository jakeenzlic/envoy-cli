package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

// NewDiffCmd returns a cobra command that shows differences between two environment vaults.
func NewDiffCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "diff <env-a> <env-b>",
		Short: "Show differences between two environment vaults",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			envA, envB := args[0], args[1]

			cfg, err := sync.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			passphraseA, err := sync.PassphraseFor(envA, cfg)
			if err != nil {
				return fmt.Errorf("passphrase for %s: %w", envA, err)
			}
			passphraseB, err := sync.PassphraseFor(envB, cfg)
			if err != nil {
				return fmt.Errorf("passphrase for %s: %w", envB, err)
			}

			vaultA, err := vault.New(vaultPathForEnv(cfg, envA), passphraseA)
			if err != nil {
				return fmt.Errorf("opening vault %s: %w", envA, err)
			}
			vaultB, err := vault.New(vaultPathForEnv(cfg, envB), passphraseB)
			if err != nil {
				return fmt.Errorf("opening vault %s: %w", envB, err)
			}

			s := sync.New(vaultA, vaultB)
			diffs, err := s.Diff()
			if err != nil {
				return fmt.Errorf("computing diff: %w", err)
			}

			if len(diffs) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No differences between %s and %s\n", envA, envB)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Differences (%s → %s):\n", envA, envB)
			for _, d := range diffs {
				fmt.Fprintln(cmd.OutOrStdout(), d)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", ".envoy.json", "Path to envoy config file")
	return cmd
}
