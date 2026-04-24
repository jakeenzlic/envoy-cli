package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

// NewRotateCmd creates a command that re-encrypts a vault with a new passphrase.
func NewRotateCmd() *cobra.Command {
	var env string

	cmd := &cobra.Command{
		Use:   "rotate",
		Short: "Re-encrypt a vault with a new passphrase",
		Long:  "Decrypts the vault using the current passphrase and re-encrypts it with a new one.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _ := cmd.Flags().GetString("config")
			cfg, err := sync.LoadConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			oldPass, err := sync.PassphraseFor(env, cfg)
			if err != nil {
				return fmt.Errorf("current passphrase: %w", err)
			}

			vaultPath := vaultPathForEnv(env, cfg)
			v, err := vault.New(vaultPath, oldPass)
			if err != nil {
				return fmt.Errorf("open vault: %w", err)
			}

			if err := v.Load(); err != nil {
				return fmt.Errorf("load vault: %w", err)
			}

			newPass, _ := cmd.Flags().GetString("new-passphrase")
			if newPass == "" {
				return fmt.Errorf("--new-passphrase must not be empty")
			}

			v2, err := vault.New(vaultPath, newPass)
			if err != nil {
				return fmt.Errorf("create rotated vault: %w", err)
			}

			for _, k := range v.Keys() {
				val, _ := v.Get(k)
				v2.Set(k, val)
			}

			if err := v2.Save(); err != nil {
				return fmt.Errorf("save rotated vault: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "vault for %q re-encrypted successfully\n", env)
			return nil
		},
	}

	cmd.Flags().StringVarP(&env, "env", "e", "local", "environment to rotate")
	cmd.Flags().String("new-passphrase", "", "new passphrase to encrypt the vault with")
	_ = cmd.MarkFlagRequired("new-passphrase")
	return cmd
}
