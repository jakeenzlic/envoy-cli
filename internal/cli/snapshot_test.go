package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicholasgasior/envoy-cli/internal/sync"
	"github.com/nicholasgasior/envoy-cli/internal/vault"
)

func setupSnapshotEnv(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "envoy.json")
	vaultPath := filepath.Join(dir, "local.vault")

	v, err := vault.New(vaultPath, "secret")
	if err != nil {
		t.Fatal(err)
	}
	_ = v.Set("DB_HOST", "localhost")
	_ = v.Set("DB_PORT", "5432")
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}

	cfg := sync.Config{
		Project:      "testapp",
		Environments: map[string]string{"local": vaultPath},
	}
	if err := sync.SaveConfig(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	os.Setenv("ENVOY_LOCAL_PASSPHRASE", "secret")
	t.Cleanup(func() { os.Unsetenv("ENVOY_LOCAL_PASSPHRASE") })
	return dir, cfgPath
}

func TestSnapshotCreatesFile(t *testing.T) {
	dir, cfgPath := setupSnapshotEnv(t)
	cmd := NewSnapshotCmd()
	cmd.Flags().String("config", cfgPath, "")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"local", "--label", "initial"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "snapshot saved") {
		t.Errorf("expected 'snapshot saved' in output, got: %s", out)
	}
	entries, err := os.ReadDir(filepath.Join(dir, ".snapshots"))
	if err != nil {
		t.Fatalf("snapshots dir missing: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 snapshot file, got %d", len(entries))
	}
	f, _ := os.Open(filepath.Join(dir, ".snapshots", entries[0].Name()))
	defer f.Close()
	var snap Snapshot
	if err := json.NewDecoder(f).Decode(&snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if snap.Label != "initial" {
		t.Errorf("expected label 'initial', got %q", snap.Label)
	}
	if len(snap.Keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(snap.Keys))
	}
}

func TestSnapshotListEmpty(t *testing.T) {
	_, cfgPath := setupSnapshotEnv(t)
	cmd := NewSnapshotListCmd()
	cmd.Flags().String("config", cfgPath, "")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no snapshots found") {
		t.Errorf("expected no snapshots message, got: %s", buf.String())
	}
}

func TestSnapshotListAfterCreate(t *testing.T) {
	_, cfgPath := setupSnapshotEnv(t)
	snap := NewSnapshotCmd()
	snap.Flags().String("config", cfgPath, "")
	snap.SetArgs([]string{"local", "--label", "v1"})
	if err := snap.Execute(); err != nil {
		t.Fatal(err)
	}
	list := NewSnapshotListCmd()
	list.Flags().String("config", cfgPath, "")
	var buf bytes.Buffer
	list.SetOut(&buf)
	if err := list.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "v1") {
		t.Errorf("expected label 'v1' in list output, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "local") {
		t.Errorf("expected env 'local' in list output, got: %s", buf.String())
	}
}
