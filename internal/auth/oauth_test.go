package auth

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestExchangeCodeRejectsNon2XXResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer server.Close()

	_, err := Exchanger{
		HTTPClient: server.Client(),
		TokenURL:   server.URL,
	}.ExchangeCode(context.Background(), OAuthConfig{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  "http://localhost:14573/callback",
	}, "code-1")
	if err == nil {
		t.Fatal("ExchangeCode() error = nil, want non-nil")
	}
}

func TestExchangeCodeRejectsMissingAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("client-1:secret-1"))
		if auth != wantAuth {
			t.Fatalf("Authorization = %q, want %q", auth, wantAuth)
		}
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Fatalf("Content-Type = %q, want application/x-www-form-urlencoded", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		values, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatalf("ParseQuery() error = %v", err)
		}
		if values.Get("code") != "code-1" {
			t.Fatalf("code = %q, want code-1", values.Get("code"))
		}
		_, _ = w.Write([]byte(`{"refresh_token":"refresh-1","token_type":"Bearer"}`))
	}))
	defer server.Close()

	_, err := Exchanger{
		HTTPClient: server.Client(),
		TokenURL:   server.URL,
	}.ExchangeCode(context.Background(), OAuthConfig{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  "http://localhost:14573/callback",
	}, "code-1")
	if err == nil {
		t.Fatal("ExchangeCode() error = nil, want non-nil")
	}
}
