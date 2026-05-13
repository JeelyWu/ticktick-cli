package domain

import "testing"

func TestReferenceErrorNotFound(t *testing.T) {
	err := &ReferenceError{
		Kind:      "task",
		Reference: "alpha",
		Matches:   nil,
	}
	if got, want := err.Error(), `task "alpha" not found`; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestReferenceErrorAmbiguous(t *testing.T) {
	err := &ReferenceError{
		Kind:      "project",
		Reference: "work",
		Matches:   []string{"work-1", "work-2"},
	}
	if got, want := err.Error(), `project "work" is ambiguous: [work-1 work-2]`; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}
