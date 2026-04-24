package sync_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/envoy-cli/internal/sync"
)

func TestSaveAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".envoy", "sync.json")

	cfg := &sync.Config{
		Environments: map[string]sync.EnvConfig{
			"local": {VaultPath: ".envoy/local.vault", PassEnvVar: "ENVOY_LOCAL_PASS"},
			"staging": {VaultPath: ".envoy/staging.vault", PassEnvVar: "ENVOY_STAGING_PASS"},
		},
	}

	if err := sync.SaveConfig(path, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	loaded, err := sync.LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if len(loaded.Environments) != 2 {
		t.Errorf("expected 2 environments, got %d", len(loaded.Environments))
	}
	if loaded.Environments["local"].VaultPath != ".envoy/local.vault" {
		t.Errorf("unexpected vault path: %s", loaded.Environments["local"].VaultPath)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := sync.LoadConfig("/nonexistent/path/sync.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(path, []byte("not json{"), 0o600)

	_, err := sync.LoadConfig(path)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestPassphraseFromEnv(t *testing.T) {
	t.Setenv("ENVOY_TEST_PASS", "supersecret")

	ec := sync.EnvConfig{PassEnvVar: "ENVOY_TEST_PASS"}
	pass, err := ec.Passphrase()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pass != "supersecret" {
		t.Errorf("expected supersecret, got %q", pass)
	}
}

func TestPassphraseEmptyEnvVar(t *testing.T) {
	ec := sync.EnvConfig{PassEnvVar: "ENVOY_MISSING_VAR_XYZ"}
	_, err := ec.Passphrase()
	if err == nil {
		t.Error("expected error for unset env var, got nil")
	}
}

var _ = json.Marshal // ensure encoding/json is used
