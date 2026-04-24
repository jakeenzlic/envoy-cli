package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"envoy-cli/internal/cli"
	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

func TestListKeysOnly(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "envoy.json")
	vaultPath := filepath.Join(dir, "local.vault")

	cfg := &sync.Config{
		Environments: map[string]sync.EnvConfig{
			"local": {VaultPath: vaultPath, PassphraseEnvVar: "TEST_PASS_LOCAL"},
		},
	}
	if err := sync.SaveConfig(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_PASS_LOCAL", "secret")

	v, err := vault.New(vaultPath, "secret")
	if err != nil {
		t.Fatal(err)
	}
	v.Set("DB_HOST", "localhost")
	v.Set("API_KEY", "abc123")
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}

	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	root := cli.NewRootCmd()
	root.AddCommand(cli.NewListCmd())

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"list", "local"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "API_KEY") {
		t.Errorf("expected API_KEY in output, got: %s", out)
	}
	if !strings.Contains(out, "DB_HOST") {
		t.Errorf("expected DB_HOST in output, got: %s", out)
	}
	if strings.Contains(out, "abc123") {
		t.Errorf("values should not appear without --values flag")
	}
}

func TestListWithValues(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "envoy.json")
	vaultPath := filepath.Join(dir, "local.vault")

	cfg := &sync.Config{
		Environments: map[string]sync.EnvConfig{
			"local": {VaultPath: vaultPath, PassphraseEnvVar: "TEST_PASS_LOCAL2"},
		},
	}
	if err := sync.SaveConfig(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_PASS_LOCAL2", "secret")

	v, err := vault.New(vaultPath, "secret")
	if err != nil {
		t.Fatal(err)
	}
	v.Set("PORT", "8080")
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}

	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	root := cli.NewRootCmd()
	root.AddCommand(cli.NewListCmd())

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"list", "local", "--values"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "PORT=8080") {
		t.Errorf("expected PORT=8080 in output, got: %s", out)
	}
}
