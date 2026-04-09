package config

import (
	"path/filepath"
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
