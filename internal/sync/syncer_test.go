package sync_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/envoy-cli/internal/sync"
	"github.com/envoy-cli/internal/vault"
)

func newTestVault(t *testing.T, name, pass string) *vault.Vault {
	t.Helper()
	dir := t.TempDir()
	v, err := vault.New(filepath.Join(dir, name+".vault"), pass)
	if err != nil {
		t.Fatalf("creating vault %s: %v", name, err)
	}
	return v
}

func TestSyncCopiesAllKeys(t *testing.T) {
	src := newTestVault(t, "src", "pass1")
	dst := newTestVault(t, "dst", "pass2")

	_ = src.Set("DB_HOST", "localhost")
	_ = src.Set("DB_PORT", "5432")

	s := sync.New(src, dst)
	synced, err := s.Sync()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(synced) != 2 {
		t.Errorf("expected 2 synced keys, got %d", len(synced))
	}

	val, err := dst.Get("DB_HOST")
	if err != nil || val != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %q (err=%v)", val, err)
	}
}

func TestDiffDetectsChanges(t *testing.T) {
	src := newTestVault(t, "src", "pass1")
	dst := newTestVault(t, "dst", "pass2")

	_ = src.Set("API_KEY", "secret")
	_ = src.Set("DEBUG", "true")
	_ = dst.Set("API_KEY", "old-secret")
	_ = dst.Set("DEBUG", "true")

	s := sync.New(src, dst)
	diffs, err := s.Diff()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diffs) != 1 || diffs[0] != "API_KEY" {
		t.Errorf("expected [API_KEY] diff, got %v", diffs)
	}
}

func TestDiffNoDifferences(t *testing.T) {
	src := newTestVault(t, "src", "pass1")
	dst := newTestVault(t, "dst", "pass2")

	_ = src.Set("FOO", "bar")
	_ = dst.Set("FOO", "bar")

	s := sync.New(src, dst)
	diffs, err := s.Diff()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diffs) != 0 {
		t.Errorf("expected no diffs, got %v", diffs)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
