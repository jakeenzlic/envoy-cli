package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"envoy-cli/internal/cli"
)

func setupHistoryDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func TestAppendAndLoadHistory(t *testing.T) {
	dir := setupHistoryDir(t)
	err := cli.AppendHistory(dir, "local", "DB_URL", "old", "new")
	if err != nil {
		t.Fatalf("AppendHistory: %v", err)
	}
	entries, err := cli.LoadHistoryExported(dir, "local")
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Key != "DB_URL" {
		t.Errorf("unexpected key: %s", entries[0].Key)
	}
	if entries[0].OldValue != "old" || entries[0].NewValue != "new" {
		t.Errorf("unexpected values: %+v", entries[0])
	}
}

func runHistoryCmd(t *testing.T, args ...string) string {
	t.Helper()
	buf := &bytes.Buffer{}
	root := &cobra.Command{Use: "envoy"}
	root.AddCommand(cli.NewHistoryCmd())
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	_ = root.Execute()
	return buf.String()
}

func TestHistoryCmdOutput(t *testing.T) {
	dir := setupHistoryDir(t)
	cfgPath := filepath.Join(dir, "envoy.json")
	_ = os.WriteFile(cfgPath, []byte(`{"project":"p","environments":{}}`), 0o600)

	_ = cli.AppendHistory(dir, "local", "KEY", "", "val")

	out := runHistoryCmd(t, "history", "--config", cfgPath, "--env", "local")
	if !strings.Contains(out, "KEY") {
		t.Errorf("expected KEY in output, got: %s", out)
	}
}

func TestHistoryCmdEnvFilter(t *testing.T) {
	dir := setupHistoryDir(t)
	cfgPath := filepath.Join(dir, "envoy.json")
	_ = os.WriteFile(cfgPath, []byte(`{"project":"p","environments":{}}`), 0o600)

	_ = cli.AppendHistory(dir, "staging", "S_KEY", "", "s")
	_ = cli.AppendHistory(dir, "local", "L_KEY", "", "l")

	out := runHistoryCmd(t, "history", "--config", cfgPath, "--env", "local")
	if strings.Contains(out, "S_KEY") {
		t.Errorf("staging key should not appear in local history")
	}
	if !strings.Contains(out, "L_KEY") {
		t.Errorf("expected L_KEY in output")
	}
}

func TestHistoryCmdNoHistory(t *testing.T) {
	dir := setupHistoryDir(t)
	cfgPath := filepath.Join(dir, "envoy.json")
	_ = os.WriteFile(cfgPath, []byte(`{"project":"p","environments":{}}`), 0o600)

	out := runHistoryCmd(t, "history", "--config", cfgPath, "--env", "local")
	if !strings.Contains(out, "No history") {
		t.Errorf("expected 'No history' message, got: %s", out)
	}
}

func TestHistoryTimestampFormat(t *testing.T) {
	dir := setupHistoryDir(t)
	_ = cli.AppendHistory(dir, "prod", "TOKEN", "a", "b")
	entries, _ := cli.LoadHistoryExported(dir, "prod")
	if len(entries) == 0 {
		t.Fatal("no entries")
	}
	if entries[0].Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
	if time.Since(entries[0].Timestamp) > 5*time.Second {
		t.Errorf("timestamp too old: %v", entries[0].Timestamp)
	}
}
