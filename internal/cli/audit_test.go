package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"envoy-cli/internal/cli"
	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

func setupAuditVault(t *testing.T) (dir string) {
	t.Helper()
	dir = t.TempDir()

	cfg := sync.Config{
		Project: "audit-test",
		Environments: []string{"local"},
		VaultDir: dir,
		Passphrases: map[string]string{"local": "AUDIT_PASS"},
	}
	t.Setenv("AUDIT_PASS", "secret")

	cfgPath := filepath.Join(dir, ".envoy.json")
	data, _ := json.Marshal(cfg)
	os.WriteFile(cfgPath, data, 0600)

	v, err := vault.New(filepath.Join(dir, "local.vault"), "secret")
	if err != nil {
		t.Fatal(err)
	}
	v.Set("DB_HOST", "localhost")
	v.Set("DB_PASS", "")
	v.Set("API_KEY", "abc123")
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestAuditShowsKeyCount(t *testing.T) {
	dir := setupAuditVault(t)
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cmd := cli.NewAuditCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--env", "local"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !containsStr(out, "Key count:") {
		t.Errorf("expected key count in output, got:\n%s", out)
	}
	if !containsStr(out, "DB_HOST") {
		t.Errorf("expected DB_HOST in output, got:\n%s", out)
	}
}

func TestAuditMarksEmptyValues(t *testing.T) {
	dir := setupAuditVault(t)
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cmd := cli.NewAuditCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--env", "local"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !containsStr(out, "empty") {
		t.Errorf("expected 'empty' status for DB_PASS, got:\n%s", out)
	}
}

func TestAuditMissingConfig(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cmd := cli.NewAuditCmd()
	cmd.SetArgs([]string{"--env", "local"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for missing config")
	}
}
