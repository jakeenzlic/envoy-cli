package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"envoy-cli/internal/cli"
	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

func setupSearchVault(t *testing.T) (dir string) {
	t.Helper()
	dir = t.TempDir()

	v, err := vault.New(filepath.Join(dir, "dev.vault"), "secret")
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}
	_ = v.Set("DATABASE_URL", "postgres://localhost/mydb")
	_ = v.Set("API_KEY", "abc123")
	_ = v.Set("REDIS_HOST", "localhost")
	_ = v.Set("DEBUG", "true")
	if err := v.Save(); err != nil {
		t.Fatalf("save vault: %v", err)
	}

	cfg := &sync.Config{
		Project: "searchtest",
		Environments: map[string]sync.EnvConfig{
			"dev": {VaultPath: filepath.Join(dir, "dev.vault")},
		},
	}
	if err := sync.SaveConfig(filepath.Join(dir, "envoy.json"), cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	t.Setenv("ENVOY_DEV_PASSPHRASE", "secret")
	return dir
}

func TestSearchMatchesKeyPrefix(t *testing.T) {
	dir := setupSearchVault(t)
	old, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer os.Chdir(old)

	var buf bytes.Buffer
	cmd := cli.NewSearchCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"dev", "DATABASE"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !containsStr(out, "DATABASE_URL") {
		t.Errorf("expected DATABASE_URL in output, got: %s", out)
	}
	if containsStr(out, "API_KEY") {
		t.Errorf("did not expect API_KEY in output, got: %s", out)
	}
}

func TestSearchNoMatches(t *testing.T) {
	dir := setupSearchVault(t)
	old, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer os.Chdir(old)

	var buf bytes.Buffer
	cmd := cli.NewSearchCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"dev", "NONEXISTENT"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsStr(buf.String(), "no matches") {
		t.Errorf("expected 'no matches' message, got: %s", buf.String())
	}
}

func TestSearchValuesFlag(t *testing.T) {
	dir := setupSearchVault(t)
	old, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer os.Chdir(old)

	var buf bytes.Buffer
	cmd := cli.NewSearchCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"dev", "localhost", "--values"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	// DATABASE_URL and REDIS_HOST both contain "localhost"
	if !containsStr(out, "DATABASE_URL") {
		t.Errorf("expected DATABASE_URL in output, got: %s", out)
	}
	if !containsStr(out, "REDIS_HOST") {
		t.Errorf("expected REDIS_HOST in output, got: %s", out)
	}
}

func TestSearchCaseSensitive(t *testing.T) {
	dir := setupSearchVault(t)
	old, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer os.Chdir(old)

	var buf bytes.Buffer
	cmd := cli.NewSearchCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"dev", "database", "--case-sensitive"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// lowercase pattern should NOT match DATABASE_URL with case-sensitive flag
	if containsStr(buf.String(), "DATABASE_URL") {
		t.Errorf("did not expect DATABASE_URL with case-sensitive lowercase search, got: %s", buf.String())
	}
}
