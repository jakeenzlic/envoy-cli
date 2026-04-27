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

func setupCompareVaults(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cfg := sync.Config{
		Project:      "compare-test",
		Environments: []string{"local", "staging"},
		VaultDir:     dir,
		Passphrases:  map[string]string{"local": "CMP_LOCAL", "staging": "CMP_STAGING"},
	}
	t.Setenv("CMP_LOCAL", "pass1")
	t.Setenv("CMP_STAGING", "pass2")

	cfgPath := filepath.Join(dir, ".envoy.json")
	data, _ := json.Marshal(cfg)
	os.WriteFile(cfgPath, data, 0600)

	v1, _ := vault.New(filepath.Join(dir, "local.vault"), "pass1")
	v1.Set("SHARED_KEY", "val1")
	v1.Set("LOCAL_ONLY", "x")
	v1.Save()

	v2, _ := vault.New(filepath.Join(dir, "staging.vault"), "pass2")
	v2.Set("SHARED_KEY", "val2")
	v2.Set("STAGING_ONLY", "y")
	v2.Save()

	return dir
}

func TestCompareShowsMissingKeys(t *testing.T) {
	dir := setupCompareVaults(t)
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cmd := cli.NewCompareCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"local", "staging"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !containsStr(out, "LOCAL_ONLY") {
		t.Errorf("expected LOCAL_ONLY in output, got:\n%s", out)
	}
	if !containsStr(out, "STAGING_ONLY") {
		t.Errorf("expected STAGING_ONLY in output, got:\n%s", out)
	}
}

func TestCompareSharedKeyNotShown(t *testing.T) {
	dir := setupCompareVaults(t)
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cmd := cli.NewCompareCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"local", "staging"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if containsStr(out, "SHARED_KEY") {
		t.Errorf("SHARED_KEY should not appear (present in both), got:\n%s", out)
	}
}

func TestCompareMissingConfig(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cmd := cli.NewCompareCmd()
	cmd.SetArgs([]string{"local", "staging"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for missing config")
	}
}
