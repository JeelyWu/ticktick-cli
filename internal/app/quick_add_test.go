package app

import "testing"

func TestParseQuickAdd(t *testing.T) {
	parsed, err := ParseQuickAdd("Write spec #Work !5 ^2026-04-10")
	if err != nil {
		t.Fatalf("ParseQuickAdd() error = %v", err)
	}
	if parsed.Title != "Write spec" {
		t.Fatalf("Title = %q, want Write spec", parsed.Title)
	}
	if parsed.ProjectRef != "Work" {
		t.Fatalf("ProjectRef = %q, want Work", parsed.ProjectRef)
	}
	if parsed.DueRaw != "2026-04-10" {
		t.Fatalf("DueRaw = %q, want 2026-04-10", parsed.DueRaw)
	}
}
