package cli_test

import (
	"path/filepath"
	"strings"
	"testing"

	"envoy-cli/internal/cli"
)

func TestHistoryLimitFlag(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "envoy.json")
	writeFile(t, cfgPath, `{"project":"p","environments":{}}`)

	for i := 0; i < 5; i++ {
		_ = cli.AppendHistory(dir, "local", "KEY", "", "v")
	}

	out := runHistoryCmd(t, "history", "--config", cfgPath, "--env", "local", "--limit", "2")
	lines := nonEmptyLines(out)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines with --limit 2, got %d: %s", len(lines), out)
	}
}

func TestHistoryPersistsAcrossAppends(t *testing.T) {
	dir := t.TempDir()
	_ = cli.AppendHistory(dir, "local", "A", "", "1")
	_ = cli.AppendHistory(dir, "local", "B", "", "2")
	_ = cli.AppendHistory(dir, "local", "C", "", "3")

	entries, err := cli.LoadHistoryExported(dir, "local")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	keys := []string{entries[0].Key, entries[1].Key, entries[2].Key}
	for _, want := range []string{"A", "B", "C"} {
		found := false
		for _, k := range keys {
			if k == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("key %s not found in history", want)
		}
	}
}

// helpers

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	import_os_writeFile(t, path, content)
}

func import_os_writeFile(t *testing.T, path, content string) {
	t.Helper()
	importOsWriteFile(t, path, []byte(content))
}

func importOsWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := writeBytes(path, data); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}

func writeBytes(path string, data []byte) error {
	import (
		"os"
	)
	return os.WriteFile(path, data, 0o600)
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}
