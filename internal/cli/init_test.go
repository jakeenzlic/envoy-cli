package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"envoy-cli/internal/sync"
)

func TestInitCreatesConfig(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	cmd := NewInitCmd()
	cmd.SetArgs([]string{"--env", "staging", "--vault-dir", "vaults"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// envoy.json must exist
	data, err := os.ReadFile("envoy.json")
	if err != nil {
		t.Fatalf("envoy.json not created: %v", err)
	}

	var cfg sync.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("envoy.json is not valid JSON: %v", err)
	}

	ec, ok := cfg.Environments["staging"]
	if !ok {
		t.Fatal("expected 'staging' environment in config")
	}
	if ec.PassphraseEnv != "ENVOY_PASSPHRASE_STAGING" {
		t.Errorf("unexpected passphrase env: %s", ec.PassphraseEnv)
	}
	if ec.VaultPath != filepath.Join("vaults", "staging.vault") {
		t.Errorf("unexpected vault path: %s", ec.VaultPath)
	}

	// vault directory must exist
	if _, err := os.Stat("vaults"); err != nil {
		t.Errorf("vault directory not created: %v", err)
	}
}

func TestInitDefaultEnv(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	cmd := NewInitCmd()
	// no flags — should default to 'local'
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	data, _ := os.ReadFile("envoy.json")
	var cfg sync.Config
	_ = json.Unmarshal(data, &cfg)

	if _, ok := cfg.Environments["local"]; !ok {
		t.Error("expected 'local' as default environment")
	}
}

func TestInitRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	// Create a pre-existing config.
	_ = os.WriteFile("envoy.json", []byte(`{}`), 0o600)

	cmd := NewInitCmd()
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error when envoy.json already exists, got nil")
	}
}

func TestToUpperSnake(t *testing.T) {
	cases := []struct{ in, want string }{
		{"local", "LOCAL"},
		{"staging", "STAGING"},
		{"production", "PRODUCTION"},
		{"ALREADY", "ALREADY"},
	}
	for _, c := range cases {
		if got := toUpperSnake(c.in); got != c.want {
			t.Errorf("toUpperSnake(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
