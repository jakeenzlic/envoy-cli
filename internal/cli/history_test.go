package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nicholasgasior/envoy-cli/internal/cli"
)

func setupHistoryDir(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "envoy.json")
	cfg := `{"environments":{"dev":{"vault":"dev.vault"}}}`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	return dir, cfgPath
}

func TestAppendAndLoadHistory(t *testing.T) {
	dir, _ := setupHistoryDir(t)

	if err := cli.AppendHistory(dir, "dev", "DB_URL", "set"); err != nil {
		t.Fatalf("AppendHistory: %v", err)
	}
	if err := cli.AppendHistory(dir, "prod", "API_KEY", "delete"); err != nil {
		t.Fatalf("AppendHistory: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".envoy_history.json"))
	if err != nil {
		t.Fatal(err)
	}
	var entries []cli.HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Key != "DB_URL" || entries[0].Action != "set" {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
	if entries[1].Environment != "prod" {
		t.Errorf("unexpected second entry env: %s", entries[1].Environment)
	}
}

func TestHistoryCmdOutput(t *testing.T) {
	dir, cfgPath := setupHistoryDir(t)

	_ = cli.AppendHistory(dir, "dev", "SECRET", "set")
	_ = cli.AppendHistory(dir, "staging", "TOKEN", "rotate")

	cmd := cli.NewHistoryCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--config", cfgPath})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "SECRET") {
		t.Errorf("expected SECRET in output, got: %s", out)
	}
	if !strings.Contains(out, "TOKEN") {
		t.Errorf("expected TOKEN in output, got: %s", out)
	}
}

func TestHistoryCmdEnvFilter(t *testing.T) {
	dir, cfgPath := setupHistoryDir(t)

	_ = cli.AppendHistory(dir, "dev", "DEV_KEY", "set")
	_ = cli.AppendHistory(dir, "prod", "PROD_KEY", "set")

	cmd := cli.NewHistoryCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--env", "dev"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "DEV_KEY") {
		t.Errorf("expected DEV_KEY in output")
	}
	if strings.Contains(out, "PROD_KEY") {
		t.Errorf("PROD_KEY should be filtered out")
	}
}

func TestHistoryCmdNoHistory(t *testing.T) {
	_, cfgPath := setupHistoryDir(t)

	cmd := cli.NewHistoryCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--config", cfgPath})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(buf.String(), "No history found") {
		t.Errorf("expected 'No history found' message")
	}
}

func TestHistoryEntryTimestamp(t *testing.T) {
	dir, _ := setupHistoryDir(t)
	before := time.Now().UTC().Add(-time.Second)
	_ = cli.AppendHistory(dir, "dev", "KEY", "import")
	after := time.Now().UTC().Add(time.Second)

	data, _ := os.ReadFile(filepath.Join(dir, ".envoy_history.json"))
	var entries []cli.HistoryEntry
	_ = json.Unmarshal(data, &entries)

	if entries[0].Timestamp.Before(before) || entries[0].Timestamp.After(after) {
		t.Errorf("timestamp out of expected range: %v", entries[0].Timestamp)
	}
}
