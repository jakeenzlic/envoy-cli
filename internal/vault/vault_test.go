package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/envoy-cli/internal/vault"
)

func TestSetAndGet(t *testing.T) {
	v := vault.New()
	v.Set("DB_HOST", "localhost")
	val, ok := v.Get("DB_HOST")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if val != "localhost" {
		t.Fatalf("expected 'localhost', got %q", val)
	}
}

func TestDelete(t *testing.T) {
	v := vault.New()
	v.Set("TOKEN", "secret")
	v.Delete("TOKEN")
	_, ok := v.Get("TOKEN")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")
	passphrase := "super-secret-passphrase"

	v := vault.New()
	v.Set("API_KEY", "abc123")
	v.Set("DB_URL", "postgres://localhost/mydb")

	if err := v.Save(path, passphrase); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := vault.Load(path, passphrase)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	for _, key := range []string{"API_KEY", "DB_URL"} {
		origVal, _ := v.Get(key)
		loadedVal, ok := loaded.Get(key)
		if !ok {
			t.Fatalf("key %q missing after load", key)
		}
		if origVal != loadedVal {
			t.Fatalf("key %q: expected %q, got %q", key, origVal, loadedVal)
		}
	}
}

func TestLoadWrongPassphrase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")

	v := vault.New()
	v.Set("SECRET", "value")
	if err := v.Save(path, "correct-pass"); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	_, err := vault.Load(path, "wrong-pass")
	if err == nil {
		t.Fatal("expected error when using wrong passphrase")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := vault.Load("/nonexistent/path.vault", "pass")
	if !os.IsNotExist(err) {
		t.Fatalf("expected not-exist error, got: %v", err)
	}
}
