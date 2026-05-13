package domain

import "testing"

func TestPriorityString(t *testing.T) {
	cases := []struct {
		p    Priority
		want string
	}{
		{PriorityNone, "none"},
		{PriorityLow, "low"},
		{PriorityMedium, "medium"},
		{PriorityHigh, "high"},
		{Priority(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.p.String(); got != tc.want {
			t.Fatalf("Priority(%d).String() = %q, want %q", tc.p, got, tc.want)
		}
	}
}

func TestTaskStatusString(t *testing.T) {
	cases := []struct {
		s    TaskStatus
		want string
	}{
		{StatusOpen, "open"},
		{StatusCompleted, "completed"},
		{TaskStatus(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.s.String(); got != tc.want {
			t.Fatalf("TaskStatus(%d).String() = %q, want %q", tc.s, got, tc.want)
		}
	}
}
