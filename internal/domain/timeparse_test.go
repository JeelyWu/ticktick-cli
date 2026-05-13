package domain

import (
	"testing"
	"time"
)

func TestParseUserTimeSupportsDateOnly(t *testing.T) {
	loc := time.FixedZone("UTC+8", 8*60*60)
	got, err := ParseUserTime("2026-04-10", loc)
	if err != nil {
		t.Fatalf("ParseUserTime() error = %v", err)
	}
	if got.Format("2006-01-02T15:04") != "2026-04-10T00:00" {
		t.Fatalf("got = %s, want midnight local date", got.Format("2006-01-02T15:04"))
	}
}

func TestParseUserTimeSupportsRFC3339(t *testing.T) {
	loc := time.FixedZone("UTC+8", 8*60*60)
	got, err := ParseUserTime("2026-04-10T14:30:00+08:00", loc)
	if err != nil {
		t.Fatalf("ParseUserTime() error = %v", err)
	}
	if got.Year() != 2026 || got.Month() != 4 || got.Day() != 10 {
		t.Fatalf("got = %s, want 2026-04-10", got.Format("2006-01-02"))
	}
	if got.Hour() != 14 || got.Minute() != 30 {
		t.Fatalf("got = %s, want 14:30", got.Format("15:04"))
	}
}
