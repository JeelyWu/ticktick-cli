package auth

import (
	"net/url"
	"testing"
)

func TestBuildAuthorizeURL(t *testing.T) {
	got := BuildAuthorizeURL(OAuthConfig{
		ClientID:    "client-1",
		RedirectURL: "http://localhost:14573/callback",
	}, "state-1")

	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	query := parsed.Query()
	if query.Get("client_id") != "client-1" {
		t.Fatalf("client_id = %q, want client-1", query.Get("client_id"))
	}
	if query.Get("redirect_uri") != "http://localhost:14573/callback" {
		t.Fatalf("redirect_uri = %q, want callback", query.Get("redirect_uri"))
	}
	if query.Get("scope") != "tasks:read tasks:write" {
		t.Fatalf("scope = %q, want tasks:read tasks:write", query.Get("scope"))
	}
}
