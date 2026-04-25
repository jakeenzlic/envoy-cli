package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"envoy-cli/internal/cli"
	"envoy-cli/internal/vault"
)

func setupDeleteVault(t *testing.T) (dir string) {
	t.Helper()
	dir = t.TempDir()

	cfgPath := filepath.Join(dir, ".envoy.json")
	cfgContent := `{"environments":{"local":{"vault":"local.vault","passphrase_env":"LOCAL_PASS"}},"default_env":"local"}`
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("LOCAL_PASS", "test-passphrase")

	vPath := filepath.Join(dir, "local.vault")
	v, err := vault.New(vPath, "test-passphrase")
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Set("DB_HOST", "localhost"); err != nil {
		t.Fatal(err)
	}
	if err := v.Set("API_KEY", "secret"); err != nil {
		t.Fatal(err)
	}
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestDeleteKeyWithYesFlag(t *testing.T) {
	dir := setupDeleteVault(t)
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cmd := cli.NewDeleteCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--yes", "--env", "local", "DB_HOST"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Deleted key") {
		t.Errorf("expected deletion message, got: %s", out.String())
	}

	v, _ := vault.New(filepath.Join(dir, "local.vault"), "test-passphrase")
	if _, err := v.Get("DB_HOST"); err == nil {
		t.Error("expected DB_HOST to be deleted from vault")
	}
	if val, err := v.Get("API_KEY"); err != nil || val != "secret" {
		t.Error("expected API_KEY to remain in vault")
	}
}

func TestDeleteNonExistentKey(t *testing.T) {
	dir := setupDeleteVault(t)
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cmd := cli.NewDeleteCmd()
	cmd.SetArgs([]string{"--yes", "--env", "local", "MISSING_KEY"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestDeleteAbortOnNoConfirmation(t *testing.T) {
	dir := setupDeleteVault(t)
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cmd := cli.NewDeleteCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetIn(strings.NewReader("n\n"))
	cmd.SetArgs([]string{"--env", "local", "DB_HOST"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Aborted") {
		t.Errorf("expected abort message, got: %s", out.String())
	}
}
