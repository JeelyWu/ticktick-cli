package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStoreRoundTrip(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "config.yaml"))
	cfg := Default()
	cfg.Output = "json"
	cfg.DefaultProject = "Inbox"

	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Output != "json" {
		t.Fatalf("loaded output = %q, want json", loaded.Output)
	}
	if loaded.DefaultProject != "Inbox" {
		t.Fatalf("loaded default project = %q, want Inbox", loaded.DefaultProject)
	}
}

func TestStoreLoadReturnsDefaultForMissingFile(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "config.yaml"))

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Output != "table" {
		t.Fatalf("loaded output = %q, want table", cfg.Output)
	}
}

func TestStoreLoadReturnsDefaultForEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	store := NewStore(path)
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.Output != "table" {
		t.Fatalf("loaded output = %q, want table", cfg.Output)
	}
}

func TestStoreLoadMergesPartialConfigWithDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("default_project: Work\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	store := NewStore(path)
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Output != "table" {
		t.Fatalf("loaded output = %q, want table", cfg.Output)
	}
	if cfg.DefaultProject != "Work" {
		t.Fatalf("loaded default project = %q, want Work", cfg.DefaultProject)
	}
}

func TestStoreLoadRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("output: json\nunknown_field: wide\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	store := NewStore(path)
	_, err := store.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want unknown field error")
	}
	if !strings.Contains(err.Error(), "field unknown_field not found") {
		t.Fatalf("Load() error = %q, want unknown field message", err.Error())
	}
}

func TestStoreSaveCreatesDirectoryWithRestrictedPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission checks skipped on windows")
	}

	path := filepath.Join(t.TempDir(), "tick", "config.yaml")
	store := NewStore(path)

	if err := store.Save(Default()); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Fatalf("directory permissions = %#o, want 0700", got)
	}
}

func TestStoreSaveCreatesFileWithRestrictedPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission checks skipped on windows")
	}

	path := filepath.Join(t.TempDir(), "tick", "config.yaml")
	store := NewStore(path)

	if err := store.Save(Default()); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("file permissions = %#o, want 0600", got)
	}
}
