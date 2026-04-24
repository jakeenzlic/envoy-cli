package sync

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds sync configuration for environments.
type Config struct {
	Environments map[string]EnvConfig `json:"environments"`
}

// EnvConfig holds per-environment vault settings.
type EnvConfig struct {
	VaultPath  string `json:"vault_path"`
	PassEnvVar string `json:"pass_env_var"`
}

// DefaultConfigPath is the default location for the sync config file.
const DefaultConfigPath = ".envoy/sync.json"

// LoadConfig reads a Config from the given JSON file path.
func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("load config: decode: %w", err)
	}
	return &cfg, nil
}

// SaveConfig writes a Config to the given JSON file path.
func SaveConfig(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("save config: mkdir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("save config: open: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		return fmt.Errorf("save config: encode: %w", err)
	}
	return nil
}

// Passphrase resolves the passphrase for an EnvConfig from the environment.
func (e *EnvConfig) Passphrase() (string, error) {
	if e.PassEnvVar == "" {
		return "", fmt.Errorf("passphrase: pass_env_var not set")
	}
	val := os.Getenv(e.PassEnvVar)
	if val == "" {
		return "", fmt.Errorf("passphrase: env var %q is empty or unset", e.PassEnvVar)
	}
	return val, nil
}
