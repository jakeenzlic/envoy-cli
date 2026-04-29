package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicholasgasior/envoy-cli/internal/sync"
	"github.com/nicholasgasior/envoy-cli/internal/vault"
)

func TestSnapshotEnvFilter(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "envoy.json")

	for _, env := range []string{"local", "staging"} {
		vp := filepath.Join(dir, env+".vault")
		v, err := vault.New(vp, "pass")
		if err != nil {
			t.Fatal(err)
		}
		_ = v.Set("KEY", env+"_value")
		if err := v.Save(); err != nil {
			t.Fatal(err)
		}
		os.Setenv("ENVOY_"+strings.ToUpper(env)+"_PASSPHRASE", "pass")
		t.Cleanup(func() { os.Unsetenv("ENVOY_" + strings.ToUpper(env) + "_PASSPHRASE") })
	}

	cfg := sync.Config{
		Project: "testapp",
		Environments: map[string]string{
			"local":   filepath.Join(dir, "local.vault"),
			"staging": filepath.Join(dir, "staging.vault"),
		},
	}
	if err := sync.SaveConfig(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}

	for _, env := range []string{"local", "staging"} {
		snap := NewSnapshotCmd()
		snap.Flags().String("config", cfgPath, "")
		snap.SetArgs([]string{env})
		if err := snap.Execute(); err != nil {
			t.Fatalf("snapshot %s: %v", env, err)
		}
	}

	list := NewSnapshotListCmd()
	list.Flags().String("config", cfgPath, "")
	list.SetArgs([]string{"--env", "local"})
	var buf bytes.Buffer
	list.SetOut(&buf)
	if err := list.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "local") {
		t.Errorf("expected 'local' in filtered output, got: %s", out)
	}
	if strings.Contains(out, "staging") {
		t.Errorf("unexpected 'staging' in filtered output: %s", out)
	}
}

func TestSnapshotPreservesAllValues(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "envoy.json")
	vp := filepath.Join(dir, "prod.vault")

	v, err := vault.New(vp, "prodpass")
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]string{"A": "1", "B": "2", "C": "3"}
	for k, val := range expected {
		_ = v.Set(k, val)
	}
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}
	os.Setenv("ENVOY_PROD_PASSPHRASE", "prodpass")
	t.Cleanup(func() { os.Unsetenv("ENVOY_PROD_PASSPHRASE") })

	cfg := sync.Config{
		Project:      "app",
		Environments: map[string]string{"prod": vp},
	}
	if err := sync.SaveConfig(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}

	snap := NewSnapshotCmd()
	snap.Flags().String("config", cfgPath, "")
	snap.SetArgs([]string{"prod"})
	if err := snap.Execute(); err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(filepath.Join(dir, ".snapshots"))
	if len(entries) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(entries))
	}
}
