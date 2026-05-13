package domain

import "testing"

func TestFocusModeString(t *testing.T) {
	cases := []struct {
		mode FocusMode
		want string
	}{
		{FocusModeTimer, "timer"},
		{FocusModePomodoro, "pomodoro"},
		{FocusMode(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.mode.String(); got != tc.want {
			t.Fatalf("FocusMode(%d).String() = %q, want %q", tc.mode, got, tc.want)
		}
	}
}

func TestFocusStatusString(t *testing.T) {
	cases := []struct {
		status FocusStatus
		want   string
	}{
		{FocusStatusActive, "active"},
		{FocusStatusCompleted, "completed"},
		{FocusStatus(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.status.String(); got != tc.want {
			t.Fatalf("FocusStatus(%d).String() = %q, want %q", tc.status, got, tc.want)
		}
	}
}
