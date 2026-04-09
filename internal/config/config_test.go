package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStoreRoundTrip(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "config.yaml"))
	cfg := Default()
	cfg.Output.Default = "json"
	cfg.Task.DefaultProject = "Inbox"

	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Output.Default != "json" {
		t.Fatalf("loaded output = %q, want json", loaded.Output.Default)
	}
	if loaded.Task.DefaultProject != "Inbox" {
		t.Fatalf("loaded default project = %q, want Inbox", loaded.Task.DefaultProject)
	}
}

func TestStoreLoadReturnsDefaultForMissingFile(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "config.yaml"))

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Output.Default != "table" {
		t.Fatalf("loaded output = %q, want table", cfg.Output.Default)
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
	if cfg.Output.Default != "table" {
		t.Fatalf("loaded output = %q, want table", cfg.Output.Default)
	}
}

func TestStoreLoadMergesPartialConfigWithDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("task:\n  default_project: Work\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	store := NewStore(path)
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Output.Default != "table" {
		t.Fatalf("loaded output = %q, want table", cfg.Output.Default)
	}
	if cfg.Task.DefaultProject != "Work" {
		t.Fatalf("loaded default project = %q, want Work", cfg.Task.DefaultProject)
	}
}

func TestStoreLoadRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("output:\n  default: json\n  format: wide\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	store := NewStore(path)
	_, err := store.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want unknown field error")
	}
	if !strings.Contains(err.Error(), "field format not found") {
		t.Fatalf("Load() error = %q, want unknown field message", err.Error())
	}
}
