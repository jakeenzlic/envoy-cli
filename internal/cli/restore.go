package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// NewRestoreCmd creates a command to restore a vault from a snapshot.
func NewRestoreCmd() *cobra.Command {
	var env string
	var yes bool

	cmd := &cobra.Command{
		Use:   "restore <snapshot-file>",
		Short: "Restore a vault from a snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			snapshotFile := args[0]

			cfg, err := loadCfgForCmd(cmd)
			if err != nil {
				return err
			}

			if env == "" {
				env = "local"
			}

			passphrase, err := passphraseForCmd(cfg, env)
			if err != nil {
				return err
			}

			// Resolve snapshot path: if not absolute, look in snapshot dir
			if !filepath.IsAbs(snapshotFile) {
				cfgPath, _ := cmd.Flags().GetString("config")
				snapshotFile = filepath.Join(snapshotDir(cfgPath), snapshotFile)
			}

			if _, err := os.Stat(snapshotFile); os.IsNotExist(err) {
				return fmt.Errorf("snapshot file not found: %s", snapshotFile)
			}

			vaultPath := vaultPathForEnv(cfg, env)

			if !yes {
				fmt.Fprintf(cmd.OutOrStdout(), "Restore vault for env %q from %s? [y/N]: ", env, filepath.Base(snapshotFile))
				var answer string
				fmt.Fscan(cmd.InOrStdin(), &answer)
				if !strings.EqualFold(strings.TrimSpace(answer), "y") {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil
				}
			}

			data, err := os.ReadFile(snapshotFile)
			if err != nil {
				return fmt.Errorf("failed to read snapshot: %w", err)
			}

			if err := os.WriteFile(vaultPath, data, 0600); err != nil {
				return fmt.Errorf("failed to restore vault: %w", err)
			}

			// Verify the restored vault is readable with the given passphrase
			v, err := openVaultForEnv(cfg, env, passphrase)
			if err != nil {
				return fmt.Errorf("restore succeeded but vault could not be opened (wrong passphrase?): %w", err)
			}

			keys := v.Keys()
			fmt.Fprintf(cmd.OutOrStdout(), "Restored vault for env %q with %d key(s) from %s\n", env, len(keys), filepath.Base(snapshotFile))
			return nil
		},
	}

	cmd.Flags().StringVarP(&env, "env", "e", "local", "environment to restore")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompt")
	return cmd
}
