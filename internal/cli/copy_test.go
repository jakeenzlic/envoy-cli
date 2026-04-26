package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"envoy-cli/internal/cli"
	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

func setupCopyVaults(t *testing.T) (cfgPath, dir string) {
	t.Helper()
	dir = t.TempDir()
	cfgPath = filepath.Join(dir, "envoy.json")

	cfg := &sync.Config{
		Project: "copy-test",
		Environments: map[string]sync.EnvConfig{
			"staging":    {VaultPath: filepath.Join(dir, "staging.vault"), PassphraseEnv: "STAGING_PASS"},
			"production": {VaultPath: filepath.Join(dir, "prod.vault"), PassphraseEnv: "PROD_PASS"},
		},
	}
	if err := sync.SaveConfig(cfgPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	t.Setenv("STAGING_PASS", "s3cr3t")
	t.Setenv("PROD_PASS", "pr0d")

	v, err := vault.New(filepath.Join(dir, "staging.vault"), "s3cr3t")
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}
	_ = v.Set("DB_URL", "postgres://staging")
	_ = v.Set("API_KEY", "abc123")
	if err := v.Save(); err != nil {
		t.Fatalf("save staging vault: %v", err)
	}
	return cfgPath, dir
}

func TestCopySingleKey(t *testing.T) {
	cfgPath, dir := setupCopyVaults(t)

	root := cli.NewRootCmd()
	root.AddCommand(cli.NewCopyCmd())
	root.SetArgs([]string{"--config", cfgPath, "copy", "DB_URL", "--from", "staging", "--to", "production"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	v, err := vault.New(filepath.Join(dir, "prod.vault"), "pr0d")
	if err != nil {
		t.Fatalf("open prod vault: %v", err)
	}
	val, ok := v.Get("DB_URL")
	if !ok {
		t.Fatal("DB_URL not found in production vault")
	}
	if val != "postgres://staging" {
		t.Errorf("expected postgres://staging, got %s", val)
	}
	if _, ok := v.Get("API_KEY"); ok {
		t.Error("API_KEY should not have been copied")
	}
}

func TestCopyAllKeys(t *testing.T) {
	cfgPath, dir := setupCopyVaults(t)

	root := cli.NewRootCmd()
	root.AddCommand(cli.NewCopyCmd())
	root.SetArgs([]string{"--config", cfgPath, "copy", "--all", "--from", "staging", "--to", "production"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	v, err := vault.New(filepath.Join(dir, "prod.vault"), "pr0d")
	if err != nil {
		t.Fatalf("open prod vault: %v", err)
	}
	for _, key := range []string{"DB_URL", "API_KEY"} {
		if _, ok := v.Get(key); !ok {
			t.Errorf("expected key %s in production vault", key)
		}
	}
}

func TestCopyNoOverwriteByDefault(t *testing.T) {
	cfgPath, dir := setupCopyVaults(t)

	// Pre-populate production with a different DB_URL
	v, _ := vault.New(filepath.Join(dir, "prod.vault"), "pr0d")
	_ = v.Set("DB_URL", "postgres://prod-original")
	_ = v.Save()

	buf := &strings.Builder{}
	root := cli.NewRootCmd()
	root.AddCommand(cli.NewCopyCmd())
	root.SetOut(buf)
	root.SetArgs([]string{"--config", cfgPath, "copy", "DB_URL", "--from", "staging", "--to", "production"})
	_ = root.Execute()

	v2, _ := vault.New(filepath.Join(dir, "prod.vault"), "pr0d")
	val, _ := v2.Get("DB_URL")
	if val != "postgres://prod-original" {
		t.Errorf("expected original value to be preserved, got %s", val)
	}
}

func TestCopySameEnvError(t *testing.T) {
	cfgPath, _ := setupCopyVaults(t)
	root := cli.NewRootCmd()
	root.AddCommand(cli.NewCopyCmd())
	root.SetArgs([]string{"--config", cfgPath, "copy", "DB_URL", "--from", "staging", "--to", "staging"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when from and to are the same")
	}
}

func TestMain_copy(m *testing.M) {
	os.Exit(m.Run())
}
