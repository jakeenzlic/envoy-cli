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

func setupDiffVaults(t *testing.T, dir string) string {
	t.Helper()

	cfg := &sync.Config{
		Project: "diff-test",
		Environments: map[string]sync.EnvConfig{
			"local": {VaultPath: filepath.Join(dir, "local.vault"), PassphraseEnvVar: "LOCAL_PASS"},
			"staging": {VaultPath: filepath.Join(dir, "staging.vault"), PassphraseEnvVar: "STAGING_PASS"},
		},
	}
	cfgPath := filepath.Join(dir, ".envoy.json")
	if err := sync.SaveConfig(cfgPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	t.Setenv("LOCAL_PASS", "local-secret")
	t.Setenv("STAGING_PASS", "staging-secret")

	vLocal, err := vault.New(cfg.Environments["local"].VaultPath, "local-secret")
	if err != nil {
		t.Fatalf("local vault: %v", err)
	}
	_ = vLocal.Set("SHARED_KEY", "same-value")
	_ = vLocal.Set("LOCAL_ONLY", "only-in-local")
	_ = vLocal.Save()

	vStaging, err := vault.New(cfg.Environments["staging"].VaultPath, "staging-secret")
	if err != nil {
		t.Fatalf("staging vault: %v", err)
	}
	_ = vStaging.Set("SHARED_KEY", "same-value")
	_ = vStaging.Set("STAGING_ONLY", "only-in-staging")
	_ = vStaging.Save()

	return cfgPath
}

func TestDiffShowsDifferences(t *testing.T) {
	dir := t.TempDir()
	cfgPath := setupDiffVaults(t, dir)

	var buf bytes.Buffer
	cmd := cli.NewDiffCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "local", "staging"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("diff cmd: %v", err)
	}

	out := buf.String()
	if !containsStr(out, "Differences") {
		t.Errorf("expected header, got: %s", out)
	}
}

func TestDiffNoDifferencesOutput(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".envoy.json")

	cfg := &sync.Config{
		Project: "nodiff-test",
		Environments: map[string]sync.EnvConfig{
			"a": {VaultPath: filepath.Join(dir, "a.vault"), PassphraseEnvVar: "PASS_A"},
			"b": {VaultPath: filepath.Join(dir, "b.vault"), PassphraseEnvVar: "PASS_B"},
		},
	}
	_ = sync.SaveConfig(cfgPath, cfg)
	t.Setenv("PASS_A", "secret-a")
	t.Setenv("PASS_B", "secret-b")

	for _, env := range []struct{ path, pass string }{
		{cfg.Environments["a"].VaultPath, "secret-a"},
		{cfg.Environments["b"].VaultPath, "secret-b"},
	} {
		v, _ := vault.New(env.path, env.pass)
		_ = v.Set("KEY", "value")
		_ = v.Save()
	}

	var buf bytes.Buffer
	cmd := cli.NewDiffCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "a", "b"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("diff cmd: %v", err)
	}

	out := buf.String()
	if !containsStr(out, "No differences") {
		t.Errorf("expected no-diff message, got: %s", out)
	}
}

func TestDiffMissingConfig(t *testing.T) {
	cmd := cli.NewDiffCmd()
	cmd.SetArgs([]string{"--config", "/nonexistent/.envoy.json", "a", "b"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for missing config")
	}
}

func init() {
	_ = os.Getenv // ensure os import is used
}
