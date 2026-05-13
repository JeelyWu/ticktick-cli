package domain

import "testing"

func TestTaskFilterStatusCodes(t *testing.T) {
	f := TaskFilter{Statuses: []TaskStatus{StatusOpen, StatusCompleted}}
	got := f.StatusCodes()
	want := []int{0, 2}
	if len(got) != len(want) {
		t.Fatalf("StatusCodes() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("StatusCodes()[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestTaskFilterStatusCodesEmpty(t *testing.T) {
	f := TaskFilter{}
	got := f.StatusCodes()
	if len(got) != 0 {
		t.Fatalf("StatusCodes() = %v, want empty", got)
	}
}

func TestTaskFilterPriorityCodes(t *testing.T) {
	f := TaskFilter{Priorities: []Priority{PriorityNone, PriorityHigh}}
	got := f.PriorityCodes()
	want := []int{0, 5}
	if len(got) != len(want) {
		t.Fatalf("PriorityCodes() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("PriorityCodes()[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestTaskFilterPriorityCodesEmpty(t *testing.T) {
	f := TaskFilter{}
	got := f.PriorityCodes()
	if len(got) != 0 {
		t.Fatalf("PriorityCodes() = %v, want empty", got)
	}
}
