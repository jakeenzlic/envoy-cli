package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"envoy-cli/internal/sync"
	"envoy-cli/internal/vault"
)

func writeEnvFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestImportCreatesKeys(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeTempConfig(t, dir, []string{"local"})

	envFile := writeEnvFile(t, dir, ".env", "APP_NAME=envoy\nSECRET_KEY=abc123\n# comment\n\nDEBUG=true\n")

	cmd := NewImportCmd()
	cmd.SetArgs([]string{"--config", cfgPath, "--env", "local", envFile})
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("import failed: %v", err)
	}

	out := buf.String()
	if !containsStr(out, "Imported 3") {
		t.Errorf("expected 3 keys imported, got: %s", out)
	}

	cfg, _ := sync.LoadConfig(cfgPath)
	passphrase, _ := sync.PassphraseFor(cfg, "local")
	v, err := vault.New(vaultPathForEnv(cfg, "local"), passphrase)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"APP_NAME", "SECRET_KEY", "DEBUG"} {
		if _, ok := v.Get(k); !ok {
			t.Errorf("expected key %q in vault", k)
		}
	}
}

func TestImportSkipsExistingWithoutOverwrite(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeTempConfig(t, dir, []string{"local"})

	cfg, _ := sync.LoadConfig(cfgPath)
	passphrase, _ := sync.PassphraseFor(cfg, "local")
	v, _ := vault.New(vaultPathForEnv(cfg, "local"), passphrase)
	v.Set("APP_NAME", "original")
	_ = v.Save()

	envFile := writeEnvFile(t, dir, ".env", "APP_NAME=new\nNEW_KEY=hello\n")

	cmd := NewImportCmd()
	cmd.SetArgs([]string{"--config", cfgPath, "--env", "local", envFile})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !containsStr(out, "skipped 1") {
		t.Errorf("expected 1 skipped, got: %s", out)
	}

	v2, _ := vault.New(vaultPathForEnv(cfg, "local"), passphrase)
	if val, _ := v2.Get("APP_NAME"); val != "original" {
		t.Errorf("expected original value to be preserved, got %q", val)
	}
}

func TestImportOverwriteFlag(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeTempConfig(t, dir, []string{"local"})

	cfg, _ := sync.LoadConfig(cfgPath)
	passphrase, _ := sync.PassphraseFor(cfg, "local")
	v, _ := vault.New(vaultPathForEnv(cfg, "local"), passphrase)
	v.Set("APP_NAME", "original")
	_ = v.Save()

	envFile := writeEnvFile(t, dir, ".env", "APP_NAME=updated\n")

	cmd := NewImportCmd()
	cmd.SetArgs([]string{"--config", cfgPath, "--env", "local", "--overwrite", envFile})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	v2, _ := vault.New(vaultPathForEnv(cfg, "local"), passphrase)
	if val, _ := v2.Get("APP_NAME"); val != "updated" {
		t.Errorf("expected updated value, got %q", val)
	}
}
