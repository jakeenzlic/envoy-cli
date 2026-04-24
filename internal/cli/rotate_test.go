package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"envoy-cli/internal/cli"
	"envoy-cli/internal/vault"
)

func TestRotateReEncryptsVault(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeTempConfig(t, dir)

	// seed vault with old passphrase
	vp := filepath.Join(dir, "local.vault")
	v, err := vault.New(vp, "old-secret")
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}
	v.Set("KEY", "value")
	if err := v.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	root := cli.NewRootCmd()
	root.AddCommand(cli.NewRotateCmd())

	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{
		"rotate",
		"--config", cfgPath,
		"--env", "local",
		"--new-passphrase", "new-secret",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("rotate: %v", err)
	}

	if !containsStr(buf.String(), "re-encrypted successfully") {
		t.Errorf("expected success message, got: %s", buf.String())
	}

	// verify old passphrase no longer works
	vOld, _ := vault.New(vp, "old-secret")
	if err := vOld.Load(); err == nil {
		t.Error("expected old passphrase to fail after rotation")
	}

	// verify new passphrase works and data is intact
	vNew, err := vault.New(vp, "new-secret")
	if err != nil {
		t.Fatalf("new vault with new pass: %v", err)
	}
	if err := vNew.Load(); err != nil {
		t.Fatalf("load with new pass: %v", err)
	}
	val, ok := vNew.Get("KEY")
	if !ok || val != "value" {
		t.Errorf("expected KEY=value after rotation, got %q ok=%v", val, ok)
	}
}

func TestRotateEmptyPassphrase(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeTempConfig(t, dir)

	vp := filepath.Join(dir, "local.vault")
	v, _ := vault.New(vp, "old-secret")
	v.Set("A", "1")
	_ = v.Save()

	root := cli.NewRootCmd()
	root.AddCommand(cli.NewRotateCmd())
	root.SetArgs([]string{
		"rotate",
		"--config", cfgPath,
		"--env", "local",
		"--new-passphrase", "",
	})

	if err := root.Execute(); err == nil {
		t.Error("expected error for empty new-passphrase")
	}
	_ = os.Remove(vp)
}
