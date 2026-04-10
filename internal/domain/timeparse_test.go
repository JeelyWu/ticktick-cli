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
