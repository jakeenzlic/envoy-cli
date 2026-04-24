package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

// NewImportCmd creates a command that imports a plain .env file into an encrypted vault.
func NewImportCmd() *cobra.Command {
	var (
		cfgPath string
		env     string
		overwrite bool
	)

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import a plain .env file into an encrypted vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			cfg, err := sync.LoadConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			passphrase, err := sync.PassphraseFor(cfg, env)
			if err != nil {
				return fmt.Errorf("passphrase: %w", err)
			}

			vaultPath := vaultPathForEnv(cfg, env)
			v, err := vault.New(vaultPath, passphrase)
			if err != nil {
				return fmt.Errorf("open vault: %w", err)
			}

			pairs, err := parseEnvFile(filePath)
			if err != nil {
				return fmt.Errorf("parse env file: %w", err)
			}

			imported := 0
			skipped := 0
			for k, val := range pairs {
				if !overwrite {
					if _, exists := v.Get(k); exists {
						skipped++
						continue
					}
				}
				v.Set(k, val)
				imported++
			}

			if err := v.Save(); err != nil {
				return fmt.Errorf("save vault: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Imported %d key(s) into [%s] vault (skipped %d).\n", imported, env, skipped)
			return nil
		},
	}

	cmd.Flags().StringVarP(&cfgPath, "config", "c", "envoy.json", "path to envoy config file")
	cmd.Flags().StringVarP(&env, "env", "e", "local", "target environment")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite existing keys")
	return cmd
}

// parseEnvFile reads KEY=VALUE pairs from a plain text .env file.
func parseEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pairs := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = strings.Trim(val, `"'`)
		pairs[key] = val
	}
	return pairs, scanner.Err()
}
