package domain

import "testing"

func TestHabitStatusString(t *testing.T) {
	cases := []struct {
		status HabitStatus
		want   string
	}{
		{HabitStatusActive, "active"},
		{HabitStatusArchived, "archived"},
		{HabitStatus(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.status.String(); got != tc.want {
			t.Fatalf("HabitStatus(%d).String() = %q, want %q", tc.status, got, tc.want)
		}
	}
}
