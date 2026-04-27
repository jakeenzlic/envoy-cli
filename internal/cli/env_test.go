package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"envoy-cli/internal/sync"
)

func writeEnvConfig(t *testing.T, dir string, cfg *sync.Config) string {
	t.Helper()
	path := filepath.Join(dir, ".envoy.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestEnvListsEnvironments(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cfg := &sync.Config{
		Project:      "myapp",
		Environments: []string{"local", "staging", "production"},
		VaultDir:     dir,
	}
	writeEnvConfig(t, dir, cfg)

	cmd := NewEnvCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	for _, env := range []string{"local", "staging", "production"} {
		if !containsStr(out, env) {
			t.Errorf("expected output to contain %q, got:\n%s", env, out)
		}
	}
}

func TestEnvShowsPaths(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cfg := &sync.Config{
		Project:      "myapp",
		Environments: []string{"local"},
		VaultDir:     dir,
	}
	writeEnvConfig(t, dir, cfg)

	cmd := NewEnvCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--paths"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !containsStr(out, "local") {
		t.Errorf("expected 'local' in output, got: %s", out)
	}
	if !containsStr(out, "missing") && !containsStr(out, "ok") {
		t.Errorf("expected vault status in output, got: %s", out)
	}
}

func TestEnvNoEnvironments(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cfg := &sync.Config{
		Project:      "empty",
		Environments: []string{},
		VaultDir:     dir,
	}
	writeEnvConfig(t, dir, cfg)

	cmd := NewEnvCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !containsStr(out, "No environments") {
		t.Errorf("expected 'No environments' message, got: %s", out)
	}
}

func TestEnvMissingConfig(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)

	cmd := NewEnvCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}
