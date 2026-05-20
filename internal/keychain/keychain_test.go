package keychain

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"
)

func TestFileStore_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(filepath.Join(dir, "keyfile"))

	key := []byte("0123456789abcdef0123456789abcdef")

	err := store.Store(key)
	if err != nil {
		t.Fatalf("Store() error: %v", err)
	}

	if !store.Exists() {
		t.Error("Exists() = false, want true")
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if !bytes.Equal(key, loaded) {
		t.Errorf("loaded key does not match stored key")
	}
}

func TestFileStore_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(filepath.Join(dir, "nonexistent"))

	_, err := store.Load()
	if !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}

	if store.Exists() {
		t.Error("Exists() = true, want false")
	}
}

func TestFileStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(filepath.Join(dir, "keyfile"))

	_ = store.Store([]byte("somekey"))

	err := store.Delete()
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	if store.Exists() {
		t.Error("Exists() = true after delete")
	}
}

func TestFileStore_DeleteNonExistent(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(filepath.Join(dir, "nonexistent"))

	err := store.Delete()
	if err != nil {
		t.Fatalf("Delete() should not error for non-existent file, got %v", err)
	}
}

// TestHashPath_Deterministic verifies the exported HashPath helper used by
// the CLI for sentinel naming returns a stable 12-char hex string.
func TestHashPath_Deterministic(t *testing.T) {
	a := HashPath("/Users/example/projects/foo")
	b := HashPath("/Users/example/projects/foo")
	if a != b {
		t.Errorf("HashPath should be deterministic, got %q vs %q", a, b)
	}
	if len(a) != 12 {
		t.Errorf("HashPath should return 12 hex chars, got len=%d (%q)", len(a), a)
	}

	c := HashPath("/Users/example/projects/bar")
	if a == c {
		t.Errorf("HashPath should differ for different inputs, both = %q", a)
	}
}

// TestNewStoreWithStatus_EnvOverride forces the file fallback via env var
// and verifies FallbackInfo is populated. This path does NOT touch the OS
// keychain so it is safe on any CI (no libsecret dependency).
func TestNewStoreWithStatus_EnvOverride(t *testing.T) {
	t.Setenv("TENE_KEYCHAIN_FALLBACK", "file")
	// HOME must resolve to a writable temp dir so the FileStore path
	// is sane (~/.tene/keyfile under temp).
	home := t.TempDir()
	t.Setenv("HOME", home)

	ks, info := NewStoreWithStatus("/some/project")
	if !info.Used {
		t.Errorf("FallbackInfo.Used should be true under env override, got false")
	}
	if info.Reason != "env_override" {
		t.Errorf("FallbackInfo.Reason = %q, want %q", info.Reason, "env_override")
	}
	if info.Path == "" {
		t.Errorf("FallbackInfo.Path should be non-empty")
	}
	if _, ok := ks.(*FileStore); !ok {
		t.Errorf("expected *FileStore under env override, got %T", ks)
	}
}
