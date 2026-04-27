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

func setupRenameVault(t *testing.T) (dir string) {
	t.Helper()
	dir = t.TempDir()

	cfg := sync.Config{
		Project: "rename-test",
		Environments: map[string]sync.EnvConfig{
			"local": {VaultPath: filepath.Join(dir, "local.vault"), PassphraseEnv: "RENAME_LOCAL_PASS"},
		},
	}
	data, _ := json.Marshal(cfg)
	os.WriteFile(filepath.Join(dir, "envoy.json"), data, 0600)

	t.Setenv("RENAME_LOCAL_PASS", "secret")

	v, err := vault.New(filepath.Join(dir, "local.vault"), "secret")
	if err != nil {
		t.Fatalf("vault.New: %v", err)
	}
	v.Set("OLD_KEY", "hello")
	if err := v.Save(); err != nil {
		t.Fatalf("vault.Save: %v", err)
	}
	return dir
}

func TestRenameKey(t *testing.T) {
	dir := setupRenameVault(t)
	old := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	var buf bytes.Buffer
	cmd := cli.NewRenameCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--env", "local", "OLD_KEY", "NEW_KEY"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !containsStr(out, "OLD_KEY") || !containsStr(out, "NEW_KEY") {
		t.Errorf("expected rename message, got: %q", out)
	}

	v, _ := vault.New(filepath.Join(dir, "local.vault"), "secret")
	if _, ok := v.Get("OLD_KEY"); ok {
		t.Error("OLD_KEY should have been removed")
	}
	if val, ok := v.Get("NEW_KEY"); !ok || val != "hello" {
		t.Errorf("NEW_KEY should be 'hello', got %q (found=%v)", val, ok)
	}
}

func TestRenameSameKey(t *testing.T) {
	dir := setupRenameVault(t)
	old := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cmd := cli.NewRenameCmd()
	cmd.SetArgs([]string{"--env", "local", "OLD_KEY", "OLD_KEY"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when old and new key are identical")
	}
}

func TestRenameMissingKey(t *testing.T) {
	dir := setupRenameVault(t)
	old := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cmd := cli.NewRenameCmd()
	cmd.SetArgs([]string{"--env", "local", "DOES_NOT_EXIST", "NEW_KEY"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for missing key")
	}
}

func TestRenameConflict(t *testing.T) {
	dir := setupRenameVault(t)
	old := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	// pre-populate the target key
	v, _ := vault.New(filepath.Join(dir, "local.vault"), "secret")
	v.Set("EXISTING_KEY", "world")
	v.Save()

	cmd := cli.NewRenameCmd()
	cmd.SetArgs([]string{"--env", "local", "OLD_KEY", "EXISTING_KEY"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when target key already exists")
	}
}
