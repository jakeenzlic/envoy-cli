package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

func setupExportVault(t *testing.T) (dir string) {
	t.Helper()
	dir = t.TempDir()

	cfg := sync.Config{
		Environments: map[string]sync.EnvConfig{
			"local": {
				VaultPath:  filepath.Join(dir, "local.vault"),
				PassEnvVar: "LOCAL_PASS",
			},
		},
	}
	cfgPath := filepath.Join(dir, "envoy.json")
	if err := sync.SaveConfig(cfgPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	t.Setenv("LOCAL_PASS", "secret")

	v, err := vault.New(cfg.Environments["local"].VaultPath, "secret")
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}
	_ = v.Set("APP_NAME", "envoy")
	_ = v.Set("DEBUG", "true")
	if err := v.Save(); err != nil {
		t.Fatalf("save vault: %v", err)
	}

	return dir
}

func TestExportToStdout(t *testing.T) {
	dir := setupExportVault(t)
	old, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer os.Chdir(old)

	root := NewRootCmd()
	root.AddCommand(NewExportCmd())

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"export", "--env", "local"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "APP_NAME=envoy") {
		t.Errorf("expected APP_NAME=envoy in output, got:\n%s", out)
	}
	if !strings.Contains(out, "DEBUG=true") {
		t.Errorf("expected DEBUG=true in output, got:\n%s", out)
	}
}

func TestExportToFile(t *testing.T) {
	dir := setupExportVault(t)
	old, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer os.Chdir(old)

	outPath := filepath.Join(dir, "exported.env")

	root := NewRootCmd()
	root.AddCommand(NewExportCmd())
	root.SetArgs([]string{"export", "--env", "local", "--output", outPath})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read exported file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "APP_NAME=envoy") {
		t.Errorf("expected APP_NAME=envoy in file, got:\n%s", content)
	}
}

func TestExportMissingConfig(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer os.Chdir(old)

	root := NewRootCmd()
	root.AddCommand(NewExportCmd())
	root.SetArgs([]string{"export", "--env", "local"})

	if err := root.Execute(); err == nil {
		t.Error("expected error for missing config, got nil")
	}
}
