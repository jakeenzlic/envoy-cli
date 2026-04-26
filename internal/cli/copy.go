package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

// NewCopyCmd creates a command that copies a key (or all keys) from one
// environment vault to another.
func NewCopyCmd() *cobra.Command {
	var overwrite bool
	var allKeys bool

	cmd := &cobra.Command{
		Use:   "copy <key> --from <env> --to <env>",
		Short: "Copy a key from one environment vault to another",
		Example: `  envoy copy DB_URL --from staging --to production
  envoy copy --all --from staging --to production`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fromEnv, _ := cmd.Flags().GetString("from")
			toEnv, _ := cmd.Flags().GetString("to")

			if fromEnv == "" || toEnv == "" {
				return fmt.Errorf("--from and --to flags are required")
			}
			if fromEnv == toEnv {
				return fmt.Errorf("--from and --to must be different environments")
			}
			if !allKeys && len(args) == 0 {
				return fmt.Errorf("provide a key name or use --all")
			}

			cfgPath, _ := cmd.Flags().GetString("config")
			cfg, err := sync.LoadConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			fromPass, err := sync.PassphraseFor(fromEnv, cfg)
			if err != nil {
				return fmt.Errorf("passphrase for %s: %w", fromEnv, err)
			}
			toPass, err := sync.PassphraseFor(toEnv, cfg)
			if err != nil {
				return fmt.Errorf("passphrase for %s: %w", toEnv, err)
			}

			src, err := vault.New(vaultPathForEnv(fromEnv, cfg), fromPass)
			if err != nil {
				return fmt.Errorf("open source vault: %w", err)
			}
			dst, err := vault.New(vaultPathForEnv(toEnv, cfg), toPass)
			if err != nil {
				return fmt.Errorf("open destination vault: %w", err)
			}

			keys := args
			if allKeys {
				keys = src.Keys().All()
			}

			copied := 0
			for _, key := range keys {
				val, ok := src.Get(key)
				if !ok {
					return fmt.Errorf("key %q not found in %s", key, fromEnv)
				}
				if _, exists := dst.Get(key); exists && !overwrite {
					fmt.Printf("skip %s (already exists; use --overwrite to replace)\n", key)
					continue
				}
				if err := dst.Set(key, val); err != nil {
					return fmt.Errorf("set %s: %w", key, err)
				}
				copied++
			}

			if err := dst.Save(); err != nil {
				return fmt.Errorf("save destination vault: %w", err)
			}
			fmt.Printf("copied %d key(s) from %s → %s\n", copied, fromEnv, toEnv)
			return nil
		},
	}

	cmd.Flags().String("from", "", "source environment")
	cmd.Flags().String("to", "", "destination environment")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite existing keys in destination")
	cmd.Flags().BoolVar(&allKeys, "all", false, "copy all keys from source")
	return cmd
}
