package cli

import (
	"fmt"

	"github.com/nicholasgasior/envoy-cli/internal/sync"
	"github.com/spf13/cobra"
)

// loadCfgForCmd reads the --config flag from the root command and loads the
// sync.Config. It is shared by commands that need config access.
func loadCfgForCmd(cmd *cobra.Command) (*sync.Config, error) {
	cfgPath, err := cmd.Root().PersistentFlags().GetString("config")
	if err != nil || cfgPath == "" {
		cfgPath = "envoy.json"
	}
	cfg, err := sync.LoadConfig(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("load config %q: %w", cfgPath, err)
	}
	return cfg, nil
}

// passphraseForCmd resolves the passphrase for a given environment from config
// or environment variables via sync.PassphraseFor.
func passphraseForCmd(cfg *sync.Config, env string) (string, error) {
	pass, err := sync.PassphraseFor(cfg, env)
	if err != nil {
		return "", fmt.Errorf("passphrase for %q: %w", env, err)
	}
	return pass, nil
}
