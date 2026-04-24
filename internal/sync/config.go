package sync

import (
	"encoding/json"
	"fmt"
	"os"
)

// EnvConfig holds per-environment vault settings.
type EnvConfig struct {
	// VaultPath is the file-system path to the encrypted vault file.
	VaultPath string `json:"vault_path"`
	// PassphraseEnv is the name of the environment variable that holds the
	// passphrase used to encrypt / decrypt this vault.
	PassphraseEnv string `json:"passphrase_env"`
}

// Config is the top-level structure stored in envoy.json.
type Config struct {
	Environments map[string]EnvConfig `json:"environments"`
}

// LoadConfig reads and parses an envoy.json file from disk.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, fmt.Errorf("config file not found: %s", path)
		}
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

// SaveConfig serialises cfg to JSON and writes it to path.
func SaveConfig(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("serialising config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// PassphraseFor resolves the passphrase for the given environment by reading
// the environment variable named in EnvConfig.PassphraseEnv.
func PassphraseFor(cfg Config, env string) (string, error) {
	ec, ok := cfg.Environments[env]
	if !ok {
		return "", fmt.Errorf("unknown environment %q", env)
	}
	if ec.PassphraseEnv == "" {
		return "", fmt.Errorf("no passphrase_env configured for environment %q", env)
	}
	val := os.Getenv(ec.PassphraseEnv)
	if val == "" {
		return "", fmt.Errorf("environment variable %s is not set or empty", ec.PassphraseEnv)
	}
	return val, nil
}
