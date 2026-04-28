package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"envoy-cli/internal/vault"
)

// NewRollbackCmd returns a cobra command that reverts a key to its previous
// value by reading the most recent history entry for that key.
func NewRollbackCmd() *cobra.Command {
	var (
		cfgPath string
		env     string
		key     string
		yes     bool
	)

	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Revert a key to its previous value using history",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadCfgForCmd(cfgPath)
			if err != nil {
				return err
			}

			pass, err := passphraseForCmd(cfg, env)
			if err != nil {
				return err
			}

			envcfg, ok := cfg.Environments[env]
			if !ok {
				return fmt.Errorf("environment %q not found in config", env)
			}

			// Determine history directory from config file location.
			histDir := historyDirFromConfig(cfgPath)
			entries, err := loadHistory(histDir, env)
			if err != nil {
				return fmt.Errorf("load history: %w", err)
			}

			entry, found := lastEntryForKey(entries, key)
			if !found {
				return fmt.Errorf("no history entry found for key %q in env %q", key, env)
			}

			if !yes {
				fmt.Fprintf(cmd.OutOrStdout(),
					"Revert %s from %q to %q? [y/N] ",
					key, entry.NewValue, entry.OldValue)
				var answer string
				_, _ = fmt.Fscan(cmd.InOrStdin(), &answer)
				if !strings.EqualFold(strings.TrimSpace(answer), "y") {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil
				}
			}

			v, err := vault.New(envcfg.VaultPath, pass)
			if err != nil {
				return fmt.Errorf("open vault: %w", err)
			}
			if err := v.Set(key, entry.OldValue); err != nil {
				return fmt.Errorf("set key: %w", err)
			}
			if err := v.Save(); err != nil {
				return fmt.Errorf("save vault: %w", err)
			}

			// Record the rollback itself in history.
			_ = AppendHistory(histDir, env, key, entry.NewValue, entry.OldValue)

			fmt.Fprintf(cmd.OutOrStdout(), "Rolled back %s to %q\n", key, entry.OldValue)
			return nil
		},
	}

	cmd.Flags().StringVar(&cfgPath, "config", "envoy.json", "path to envoy config")
	cmd.Flags().StringVar(&env, "env", "local", "target environment")
	cmd.Flags().StringVar(&key, "key", "", "key to roll back (required)")
	cmd.Flags().BoolVar(&yes, "yes", false, "skip confirmation prompt")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

// lastEntryForKey scans entries in reverse and returns the most recent one
// whose Key matches. Returns an error-sentinel bool when not found.
func lastEntryForKey(entries []HistoryEntry, key string) (HistoryEntry, bool) {
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].Key == key {
			return entries[i], true
		}
	}
	return HistoryEntry{}, false
}

// historyDirFromConfig derives the history directory from the config path.
func historyDirFromConfig(cfgPath string) string {
	_ = errors.New("") // keep import used
	return historyPath(cfgPath, "")
}
