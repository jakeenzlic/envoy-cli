package cli_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"envoy-cli/internal/cli"
	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

// TestRenamePreservesOtherKeys ensures unrelated keys are untouched after a rename.
func TestRenamePreservesOtherKeys(t *testing.T) {
	dir := t.TempDir()

	cfg := sync.Config{
		Project: "integ-rename",
		Environments: map[string]sync.EnvConfig{
			"local": {
				VaultPath:     filepath.Join(dir, "local.vault"),
				PassphraseEnv: "INTEG_RENAME_PASS",
			},
		},
	}
	data, _ := json.Marshal(cfg)
	os.WriteFile(filepath.Join(dir, "envoy.json"), data, 0600)
	t.Setenv("INTEG_RENAME_PASS", "passw0rd")

	v, err := vault.New(filepath.Join(dir, "local.vault"), "passw0rd")
	if err != nil {
		t.Fatalf("vault.New: %v", err)
	}
	v.Set("ALPHA", "one")
	v.Set("BETA", "two")
	v.Set("GAMMA", "three")
	if err := v.Save(); err != nil {
		t.Fatalf("vault.Save: %v", err)
	}

	old := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cmd := cli.NewRenameCmd()
	cmd.SetArgs([]string{"--env", "local", "BETA", "BETA_RENAMED"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("rename failed: %v", err)
	}

	v2, err := vault.New(filepath.Join(dir, "local.vault"), "passw0rd")
	if err != nil {
		t.Fatalf("reopen vault: %v", err)
	}

	for _, key := range []string{"ALPHA", "GAMMA"} {
		if _, ok := v2.Get(key); !ok {
			t.Errorf("key %q should still exist after rename", key)
		}
	}
	if _, ok := v2.Get("BETA"); ok {
		t.Error("BETA should have been removed")
	}
	if val, ok := v2.Get("BETA_RENAMED"); !ok || val != "two" {
		t.Errorf("BETA_RENAMED should be 'two', got %q (found=%v)", val, ok)
	}
}
