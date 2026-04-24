package cli_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/envoy-cli/internal/cli"
	"github.com/envoy-cli/internal/sync"
)

func writeTempConfig(t *testing.T, dir string) string {
	t.Helper()
	cfg := sync.Config{
		Passphrase:   "test-secret",
		Environments: map[string]string{"local": filepath.Join(dir, "local.vault")},
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	cfgPath := filepath.Join(dir, "envoy.json")
	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return cfgPath
}

func TestPushThenPull(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeTempConfig(t, dir)

	envSrc := filepath.Join(dir, "source.env")
	if err := os.WriteFile(envSrc, []byte("FOO=bar\nBAZ=qux\n"), 0600); err != nil {
		t.Fatalf("write source env: %v", err)
	}

	pushCmd := cli.NewPushCmd()
	pushCmd.SetArgs([]string{"--file", envSrc, "--env", "local", "--config", cfgPath})
	if err := pushCmd.Execute(); err != nil {
		t.Fatalf("push: %v", err)
	}

	envDst := filepath.Join(dir, "pulled.env")
	pullCmd := cli.NewPullCmd()
	pullCmd.SetArgs([]string{"--file", envDst, "--env", "local", "--config", cfgPath})
	if err := pullCmd.Execute(); err != nil {
		t.Fatalf("pull: %v", err)
	}

	got, err := os.ReadFile(envDst)
	if err != nil {
		t.Fatalf("read pulled file: %v", err)
	}

	for _, want := range []string{"FOO=bar", "BAZ=qux"} {
		if !contains(string(got), want) {
			t.Errorf("pulled file missing %q; got:\n%s", want, got)
		}
	}
}

func TestPullNoOverwrite(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeTempConfig(t, dir)

	existing := filepath.Join(dir, "existing.env")
	if err := os.WriteFile(existing, []byte("EXISTING=1\n"), 0600); err != nil {
		t.Fatal(err)
	}

	pullCmd := cli.NewPullCmd()
	pullCmd.SetArgs([]string{"--file", existing, "--env", "local", "--config", cfgPath})
	if err := pullCmd.Execute(); err == nil {
		t.Fatal("expected error when pulling to existing file without --overwrite")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub ||
		len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
