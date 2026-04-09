package domain

import "testing"

func TestPriorityString(t *testing.T) {
	if got := PriorityHigh.String(); got != "high" {
		t.Fatalf("PriorityHigh.String() = %q, want high", got)
	}
	if got := Priority(99).String(); got != "unknown" {
		t.Fatalf("Priority(99).String() = %q, want unknown", got)
	}
}

func TestTaskStatusString(t *testing.T) {
	if got := StatusOpen.String(); got != "open" {
		t.Fatalf("StatusOpen.String() = %q, want open", got)
	}
	if got := StatusCompleted.String(); got != "completed" {
		t.Fatalf("StatusCompleted.String() = %q, want completed", got)
	}
}
