package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSetAndGetValue(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeTempConfig(t, dir, []string{"local"})
	_ = os.Setenv("ENVOY_LOCAL_PASSPHRASE", "test-passphrase")
	t.Cleanup(func() { os.Unsetenv("ENVOY_LOCAL_PASSPHRASE") })

	setCmd := NewSetCmd()
	setCmd.SetArgs([]string{"--config", cfgPath, "--env", "local", "MY_KEY=hello_world"})
	var setBuf bytes.Buffer
	setCmd.SetOut(&setBuf)
	if err := setCmd.Execute(); err != nil {
		t.Fatalf("set command failed: %v", err)
	}
	if !containsStr(setBuf.String(), "MY_KEY") {
		t.Errorf("expected output to mention MY_KEY, got: %s", setBuf.String())
	}

	getCmd := NewGetCmd()
	getCmd.SetArgs([]string{"--config", cfgPath, "--env", "local", "MY_KEY"})
	var getBuf bytes.Buffer
	getCmd.SetOut(&getBuf)
	if err := getCmd.Execute(); err != nil {
		t.Fatalf("get command failed: %v", err)
	}
	if !containsStr(getBuf.String(), "hello_world") {
		t.Errorf("expected value 'hello_world', got: %s", getBuf.String())
	}
}

func TestSetInvalidFormat(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeTempConfig(t, dir, []string{"local"})
	_ = os.Setenv("ENVOY_LOCAL_PASSPHRASE", "test-passphrase")
	t.Cleanup(func() { os.Unsetenv("ENVOY_LOCAL_PASSPHRASE") })

	setCmd := NewSetCmd()
	setCmd.SetArgs([]string{"--config", cfgPath, "--env", "local", "BADFORMAT"})
	setCmd.SetOut(&bytes.Buffer{})
	if err := setCmd.Execute(); err == nil {
		t.Error("expected error for invalid KEY=VALUE format, got nil")
	}
}

func TestGetMissingKey(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeTempConfig(t, dir, []string{"local"})
	_ = os.Setenv("ENVOY_LOCAL_PASSPHRASE", "test-passphrase")
	t.Cleanup(func() { os.Unsetenv("ENVOY_LOCAL_PASSPHRASE") })

	// Ensure vault file exists by setting a different key first
	setCmd := NewSetCmd()
	setCmd.SetArgs([]string{"--config", cfgPath, "--env", "local", "OTHER=value"})
	setCmd.SetOut(&bytes.Buffer{})
	_ = setCmd.Execute()

	getCmd := NewGetCmd()
	getCmd.SetArgs([]string{"--config", cfgPath, "--env", "local", "NONEXISTENT"})
	getCmd.SetOut(&bytes.Buffer{})
	if err := getCmd.Execute(); err == nil {
		t.Error("expected error for missing key, got nil")
	}
}

func TestSetMissingConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "nonexistent.json")

	setCmd := NewSetCmd()
	setCmd.SetArgs([]string{"--config", cfgPath, "--env", "local", "KEY=VALUE"})
	setCmd.SetOut(&bytes.Buffer{})
	if err := setCmd.Execute(); err == nil {
		t.Error("expected error for missing config, got nil")
	}
}
