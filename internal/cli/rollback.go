package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// NewRollbackCmd returns a command that restores a vault to a previously
// recorded history snapshot. It reads the history log, lets the user
// pick a revision by index, and re-applies every key/value pair that
// was captured at that point in time.
func NewRollbackCmd() *cobra.Command {
	var (
		envName string
		yesFlag bool
	)

	cmd := &cobra.Command{
		Use:   "rollback <revision>",
		Short: "Restore a vault to a previous history snapshot",
		Long: `Rollback replaces the current vault contents with the key/value
pairs captured in the specified history revision.

Revision numbers are shown by the 'history' command (0 = most recent).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			revStr := args[0]
			revIdx, err := strconv.Atoi(revStr)
			if err != nil || revIdx < 0 {
				return fmt.Errorf("revision must be a non-negative integer, got %q", revStr)
			}

			cfg, err := loadCfgForCmd(cmd)
			if err != nil {
				return err
			}

			passphrase, err := passphraseForCmd(cmd, cfg, envName)
			if err != nil {
				return err
			}

			// Load history entries for the target environment.
			hPath := historyPath(cmd, envName)
			entries, err := loadHistory(hPath)
			if err != nil {
				return fmt.Errorf("could not load history: %w", err)
			}
			if len(entries) == 0 {
				return fmt.Errorf("no history found for environment %q", envName)
			}

			if revIdx >= len(entries) {
				return fmt.Errorf("revision %d out of range (history has %d entries)", revIdx, len(entries))
			}

			// History is stored newest-first; index 0 is the latest snapshot.
			target := entries[revIdx]
			if len(target.Keys) == 0 {
				return fmt.Errorf("revision %d contains no key snapshot; cannot rollback", revIdx)
			}

			if !yesFlag {
				fmt.Fprintf(cmd.OutOrStdout(),
					"Rolling back %q to revision %d (%s) will overwrite %d key(s).\nContinue? [y/N] ",
					envName, revIdx, target.Timestamp, len(target.Keys))
				var answer string
				fmt.Fscan(os.Stdin, &answer)
				if answer != "y" && answer != "Y" {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil
				}
			}

			// Open (or create) the vault for the target environment.
			vaultPath := vaultPathForEnv(cfg, envName)
			v, err := openVaultForEnv(vaultPath, passphrase)
			if err != nil {
				return fmt.Errorf("could not open vault: %w", err)
			}

			// Overwrite every key captured in the snapshot.
			for k, val := range target.Keys {
				if err := v.Set(k, val); err != nil {
					return fmt.Errorf("failed to set %q: %w", k, err)
				}
			}

			if err := v.Save(vaultPath, passphrase); err != nil {
				return fmt.Errorf("failed to save vault: %w", err)
			}

			// Record the rollback action in history.
			_ = AppendHistory(hPath, "rollback", envName, target.Keys)

			fmt.Fprintf(cmd.OutOrStdout(),
				"Rolled back %q to revision %d (%s) — %d key(s) restored.\n",
				envName, revIdx, target.Timestamp, len(target.Keys))
			return nil
		},
	}

	cmd.Flags().StringVarP(&envName, "env", "e", "local", "Target environment")
	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}
