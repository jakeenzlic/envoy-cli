package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envoy-cli/internal/sync"
	"github.com/nicholasgasior/envoy-cli/internal/vault"
)

func setupRestoreEnv(t *testing.T) (cfgPath, snapDir string) {
	t.Helper()
	dir := t.TempDir()
	cfgPath = filepath.Join(dir, "envoy.json")
	snapDir = filepath.Join(dir, ".snapshots")

	vaultPath := filepath.Join(dir, "local.vault")
	cfg := sync.Config{
		Environments: map[string]sync.EnvConfig{
			"local": {VaultPath: vaultPath},
		},
		Passphrases: map[string]string{
			"local": "secret",
		},
	}
	b, _ := json.Marshal(cfg)
	os.WriteFile(cfgPath, b, 0600)

	v, _ := vault.New(vaultPath, "secret")
	v.Set("DB_HOST", "localhost")
	v.Set("DB_PORT", "5432")
	v.Save()

	os.MkdirAll(snapDir, 0755)
	return cfgPath, snapDir
}

func runRestoreCmd(t *testing.T, cfgPath string, args ...string) (string, error) {
	t.Helper()
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	allArgs := append([]string{"--config", cfgPath, "restore"}, args...)
	cmd.SetArgs(allArgs)
	err := cmd.Execute()
	return buf.String(), err
}

func TestRestoreFromSnapshot(t *testing.T) {
	cfgPath, snapDir := setupRestoreEnv(t)

	// Create a snapshot to restore from
	dir := filepath.Dir(cfgPath)
	vaultPath := filepath.Join(dir, "local.vault")
	data, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("failed to read vault: %v", err)
	}
	snapshotFile := filepath.Join(snapDir, "local_snapshot.vault")
	os.WriteFile(snapshotFile, data, 0600)

	out, err := runRestoreCmd(t, cfgPath, "--env", "local", "--yes", snapshotFile)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}
	if !containsStr(out, "Restored vault") {
		t.Errorf("expected success message, got: %s", out)
	}
	if !containsStr(out, "2 key(s)") {
		t.Errorf("expected key count in output, got: %s", out)
	}
}

func TestRestoreMissingSnapshot(t *testing.T) {
	cfgPath, _ := setupRestoreEnv(t)

	_, err := runRestoreCmd(t, cfgPath, "--env", "local", "--yes", "nonexistent.vault")
	if err == nil {
		t.Fatal("expected error for missing snapshot")
	}
}

func TestRestoreAbortOnNoConfirmation(t *testing.T) {
	cfgPath, snapDir := setupRestoreEnv(t)
	dir := filepath.Dir(cfgPath)
	vaultPath := filepath.Join(dir, "local.vault")
	data, _ := os.ReadFile(vaultPath)
	snapshotFile := filepath.Join(snapDir, "local_snapshot.vault")
	os.WriteFile(snapshotFile, data, 0600)

	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetIn(bytes.NewBufferString("n\n"))
	cmd.SetArgs([]string{"--config", cfgPath, "restore", "--env", "local", snapshotFile})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsStr(buf.String(), "Aborted") {
		t.Errorf("expected abort message, got: %s", buf.String())
	}
}
