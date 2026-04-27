package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionDefaultOutput(t *testing.T) {
	// Override build-time vars for deterministic test output.
	Version = "1.2.3"
	Commit = "abc1234"
	BuildDate = "2024-06-01"

	cmd := NewVersionCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "1.2.3") {
		t.Errorf("expected version in output, got: %s", out)
	}
	if !strings.Contains(out, "abc1234") {
		t.Errorf("expected commit in output, got: %s", out)
	}
	if !strings.Contains(out, "2024-06-01") {
		t.Errorf("expected build date in output, got: %s", out)
	}
}

func TestVersionShortFlag(t *testing.T) {
	Version = "1.2.3"
	Commit = "abc1234"
	BuildDate = "2024-06-01"

	cmd := NewVersionCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--short"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := strings.TrimSpace(buf.String())
	if out != "1.2.3" {
		t.Errorf("expected only version number %q, got %q", "1.2.3", out)
	}
	if strings.Contains(out, "commit") {
		t.Errorf("short flag should not include commit, got: %s", out)
	}
}

func TestVersionDevDefault(t *testing.T) {
	Version = "dev"
	Commit = "none"
	BuildDate = "unknown"

	cmd := NewVersionCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "dev") {
		t.Errorf("expected 'dev' version in output, got: %s", out)
	}
}
