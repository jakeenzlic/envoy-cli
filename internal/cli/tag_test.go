package cli_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envoy-cli/internal/sync"
	"github.com/nicholasgasior/envoy-cli/internal/vault"
)

func setupTagVault(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	passphrase := "tag-secret"
	vaultFile := filepath.Join(dir, "local.vault")

	v, err := vault.New(vaultFile, passphrase)
	if err != nil {
		t.Fatalf("vault.New: %v", err)
	}
	keys := map[string]string{
		"DB_HOST":     "localhost",
		"DB_PORT":     "5432",
		"APP_NAME":    "envoy",
		"APP_VERSION": "1.0",
		"STANDALONE":  "yes",
	}
	for k, val := range keys {
		if err := v.Set(k, val); err != nil {
			t.Fatalf("Set %s: %v", k, err)
		}
	}
	if err := v.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	cfgFile := filepath.Join(dir, "envoy.json")
	cfg := &sync.Config{
		Project: "tag-test",
		Environments: map[string]sync.EnvConfig{
			"local": {VaultPath: vaultFile},
		},
		Passphrases: map[string]string{
			"local": passphrase,
		},
	}
	if err := sync.SaveConfig(cfgFile, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	return dir, cfgFile
}

func TestTagListsAllTags(t *testing.T) {
	_, cfgFile := setupTagVault(t)

	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--config", cfgFile, "tag"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	out := buf.String()
	if !containsStr(out, "DB") {
		t.Errorf("expected tag DB in output, got:\n%s", out)
	}
	if !containsStr(out, "APP") {
		t.Errorf("expected tag APP in output, got:\n%s", out)
	}
}

func TestTagFiltersByPrefix(t *testing.T) {
	_, cfgFile := setupTagVault(t)

	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--config", cfgFile, "tag", "DB"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	out := buf.String()
	if !containsStr(out, "DB_HOST") {
		t.Errorf("expected DB_HOST, got:\n%s", out)
	}
	if !containsStr(out, "DB_PORT") {
		t.Errorf("expected DB_PORT, got:\n%s", out)
	}
	if containsStr(out, "APP_NAME") {
		t.Errorf("unexpected APP_NAME in DB tag output:\n%s", out)
	}
}

func TestTagNoMatchReturnsMessage(t *testing.T) {
	_, cfgFile := setupTagVault(t)

	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--config", cfgFile, "tag", "REDIS"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	out := buf.String()
	if !containsStr(out, "no keys found") {
		t.Errorf("expected 'no keys found' message, got:\n%s", out)
	}
}
