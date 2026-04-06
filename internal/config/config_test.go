package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Version != 1 {
		t.Errorf("Version = %d, want 1", cfg.Version)
	}
	if cfg.Analytics.SyncAttempts != 0 {
		t.Errorf("SyncAttempts = %d, want 0", cfg.Analytics.SyncAttempts)
	}
	if cfg.Analytics.LastSyncAttempt != nil {
		t.Error("LastSyncAttempt should be nil")
	}
	if !cfg.Preferences.Color {
		t.Error("Color should be true")
	}
	if !cfg.Preferences.AutoKeychain {
		t.Error("AutoKeychain should be true")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg := DefaultConfig()
	cfg.Analytics.SyncAttempts = 3

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.Analytics.SyncAttempts != 3 {
		t.Errorf("SyncAttempts = %d, want 3", loaded.Analytics.SyncAttempts)
	}

	// Check file permissions
	path := filepath.Join(tmpHome, ".tene", "config.json")
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Errorf("permission = %o, want 0600", info.Mode().Perm())
	}
}

func TestLoad_FileNotExists(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Version != 1 {
		t.Errorf("Version = %d, want 1 (default)", cfg.Version)
	}
}

func TestIncrementSyncAttempts(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if err := IncrementSyncAttempts(); err != nil {
		t.Fatalf("IncrementSyncAttempts() error: %v", err)
	}
	if err := IncrementSyncAttempts(); err != nil {
		t.Fatalf("IncrementSyncAttempts() error: %v", err)
	}

	cfg, _ := Load()
	if cfg.Analytics.SyncAttempts != 2 {
		t.Errorf("SyncAttempts = %d, want 2", cfg.Analytics.SyncAttempts)
	}
	if cfg.Analytics.LastSyncAttempt == nil {
		t.Error("LastSyncAttempt should not be nil")
	}
}

func TestEnsureConfigDir(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() error: %v", err)
	}

	dir := filepath.Join(tmpHome, ".tene")
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}
