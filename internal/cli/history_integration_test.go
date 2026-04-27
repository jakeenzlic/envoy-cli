package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicholasgasior/envoy-cli/internal/cli"
)

// TestHistoryLimitFlag verifies that --limit caps the number of displayed entries.
func TestHistoryLimitFlag(t *testing.T) {
	dir, cfgPath := setupHistoryDir(t)

	keys := []string{"A", "B", "C", "D", "E"}
	for _, k := range keys {
		if err := cli.AppendHistory(dir, "dev", k, "set"); err != nil {
			t.Fatal(err)
		}
	}

	cmd := cli.NewHistoryCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--limit", "2"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines with --limit 2, got %d: %s", len(lines), buf.String())
	}
	// Should show the most recent entries (D and E)
	if !strings.Contains(buf.String(), "D") || !strings.Contains(buf.String(), "E") {
		t.Errorf("expected last 2 keys D and E, got: %s", buf.String())
	}
}

// TestHistoryPersistsAcrossAppends verifies that entries accumulate correctly.
func TestHistoryPersistsAcrossAppends(t *testing.T) {
	dir, _ := setupHistoryDir(t)

	for i := 0; i < 5; i++ {
		if err := cli.AppendHistory(dir, "dev", "KEY", "set"); err != nil {
			t.Fatal(err)
		}
	}

	data, err := os.ReadFile(filepath.Join(dir, ".envoy_history.json"))
	if err != nil {
		t.Fatal(err)
	}
	var entries []cli.HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 5 {
		t.Errorf("expected 5 accumulated entries, got %d", len(entries))
	}
}
