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

func TestFocusModeAPIType(t *testing.T) {
	cases := []struct {
		mode FocusMode
		want int
	}{
		{FocusModeTimer, 1},
		{FocusModePomodoro, 0},
		{FocusMode(99), 1},
	}
	for _, tc := range cases {
		if got := tc.mode.APIType(); got != tc.want {
			t.Fatalf("FocusMode(%d).APIType() = %d, want %d", tc.mode, got, tc.want)
		}
	}
}

func TestFocusModeFromAPIType(t *testing.T) {
	cases := []struct {
		apiType int
		want    FocusMode
	}{
		{0, FocusModePomodoro},
		{1, FocusModeTimer},
		{99, FocusModeTimer},
	}
	for _, tc := range cases {
		if got := FocusModeFromAPIType(tc.apiType); got != tc.want {
			t.Fatalf("FocusModeFromAPIType(%d) = %v, want %v", tc.apiType, got, tc.want)
		}
	}
}
