package vault_test

import (
	"path/filepath"
	"testing"

	"envoy-cli/internal/vault"
)

func TestKeysReturnsAllKeys(t *testing.T) {
	dir := t.TempDir()
	v, err := vault.New(filepath.Join(dir, "test.vault"), "pass")
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}

	v.Set("ZEBRA", "z")
	v.Set("APPLE", "a")
	v.Set("MANGO", "m")

	keys := v.Keys()
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}

	// must be sorted
	expected := []string{"APPLE", "MANGO", "ZEBRA"}
	for i, k := range keys {
		if k != expected[i] {
			t.Errorf("keys[%d] = %q, want %q", i, k, expected[i])
		}
	}
}

func TestLenReturnsCount(t *testing.T) {
	dir := t.TempDir()
	v, err := vault.New(filepath.Join(dir, "test.vault"), "pass")
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}

	if v.Len() != 0 {
		t.Errorf("expected 0, got %d", v.Len())
	}

	v.Set("A", "1")
	v.Set("B", "2")

	if v.Len() != 2 {
		t.Errorf("expected 2, got %d", v.Len())
	}

	v.Delete("A")

	if v.Len() != 1 {
		t.Errorf("expected 1 after delete, got %d", v.Len())
	}
}

func TestKeysEmptyVault(t *testing.T) {
	dir := t.TempDir()
	v, err := vault.New(filepath.Join(dir, "empty.vault"), "pass")
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}

	keys := v.Keys()
	if len(keys) != 0 {
		t.Errorf("expected empty keys slice, got %v", keys)
	}
}
