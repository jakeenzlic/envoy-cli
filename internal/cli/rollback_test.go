package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"envoy-cli/internal/cli"
	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

func setupRollbackEnv(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "local.vault")
	v, err := vault.New(vaultPath, "secret")
	if err != nil {
		t.Fatalf("vault.New: %v", err)
	}
	_ = v.Set("KEY1", "value1")
	_ = v.Set("KEY2", "value2")
	if err := v.Save(); err != nil {
		t.Fatalf("vault.Save: %v", err)
	}

	cfg := sync.Config{
		Project: "rollback-proj",
		Environments: map[string]sync.EnvConfig{
			"local": {VaultPath: vaultPath, PassphraseEnv: "TEST_PASS"},
		},
	}
	cfgPath := filepath.Join(dir, "envoy.json")
	if err := sync.SaveConfig(cfgPath, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	t.Setenv("TEST_PASS", "secret")
	return dir, cfgPath
}

func writeHistory(t *testing.T, dir, env string, entries []cli.HistoryEntry) {
	t.Helper()
	hp := cli.HistoryPath(dir, env)
	_ = os.MkdirAll(filepath.Dir(hp), 0o755)
	f, err := os.Create(hp)
	if err != nil {
		t.Fatalf("create history: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range entries {
		_ = enc.Encode(e)
	}
}

func runRollbackCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	buf := &bytes.Buffer{}
	root := &cobra.Command{Use: "envoy"}
	root.AddCommand(cli.NewRollbackCmd())
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestRollbackRestoresKey(t *testing.T) {
	dir, cfgPath := setupRollbackEnv(t)
	entries := []cli.HistoryEntry{
		{Timestamp: time.Now().Add(-2 * time.Minute), Env: "local", Key: "KEY1", OldValue: "original", NewValue: "value1"},
	}
	writeHistory(t, dir, "local", entries)

	_, err := runRollbackCmd(t, "rollback", "--config", cfgPath, "--env", "local", "--key", "KEY1", "--yes")
	if err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	v, _ := vault.New(filepath.Join(dir, "local.vault"), "secret")
	val, _ := v.Get("KEY1")
	if val != "original" {
		t.Errorf("expected 'original', got %q", val)
	}
}

func TestRollbackNoHistoryEntry(t *testing.T) {
	_, cfgPath := setupRollbackEnv(t)
	out, err := runRollbackCmd(t, "rollback", "--config", cfgPath, "--env", "local", "--key", "MISSING", "--yes")
	if err == nil {
		t.Errorf("expected error, got nil; output: %s", out)
	}
}

func TestRollbackMissingConfig(t *testing.T) {
	_, err := runRollbackCmd(t, "rollback", "--config", "/nonexistent/envoy.json", "--env", "local", "--key", "KEY1", "--yes")
	if err == nil {
		t.Error("expected error for missing config")
	}
}
