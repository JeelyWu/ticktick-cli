package endpoint

import "testing"

func TestForRegionDefaultsToTickTick(t *testing.T) {
	endpoints, err := ForRegion("")
	if err != nil {
		t.Fatalf("ForRegion(\"\") error = %v", err)
	}
	if got, want := endpoints.Region, RegionTickTick; got != want {
		t.Fatalf("Region = %q, want %q", got, want)
	}
	if got, want := endpoints.AuthorizeURL, "https://ticktick.com/oauth/authorize"; got != want {
		t.Fatalf("AuthorizeURL = %q, want %q", got, want)
	}
	if got, want := endpoints.TokenURL, "https://ticktick.com/oauth/token"; got != want {
		t.Fatalf("TokenURL = %q, want %q", got, want)
	}
	if got, want := endpoints.APIBaseURL, "https://api.ticktick.com"; got != want {
		t.Fatalf("APIBaseURL = %q, want %q", got, want)
	}
}

func TestForRegionReturnsDida365Endpoints(t *testing.T) {
	endpoints, err := ForRegion("dida365")
	if err != nil {
		t.Fatalf("ForRegion(dida365) error = %v", err)
	}
	if got, want := endpoints.Region, RegionDida365; got != want {
		t.Fatalf("Region = %q, want %q", got, want)
	}
	if got, want := endpoints.AuthorizeURL, "https://dida365.com/oauth/authorize"; got != want {
		t.Fatalf("AuthorizeURL = %q, want %q", got, want)
	}
	if got, want := endpoints.TokenURL, "https://dida365.com/oauth/token"; got != want {
		t.Fatalf("TokenURL = %q, want %q", got, want)
	}
	if got, want := endpoints.APIBaseURL, "https://api.dida365.com"; got != want {
		t.Fatalf("APIBaseURL = %q, want %q", got, want)
	}
}

func TestForRegionRejectsUnsupportedValue(t *testing.T) {
	_, err := ForRegion("unknown")
	if err == nil {
		t.Fatal("ForRegion(unknown) error = nil, want non-nil")
	}
}
