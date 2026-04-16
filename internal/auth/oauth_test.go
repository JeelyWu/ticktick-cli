package auth

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestBuildAuthorizeURL(t *testing.T) {
	got := BuildAuthorizeURL("https://ticktick.com/oauth/authorize", OAuthConfig{
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

func TestExchangeCodeStoresExpiryMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"access-1","refresh_token":"refresh-1","token_type":"Bearer","expires_in":7200}`))
	}))
	defer server.Close()

	now := time.Unix(1_776_351_966, 0).UTC()
	token, err := Exchanger{
		HTTPClient: server.Client(),
		TokenURL:   server.URL,
		Now:        func() time.Time { return now },
	}.ExchangeCode(context.Background(), OAuthConfig{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  "http://localhost:14573/callback",
	}, "code-1")
	if err != nil {
		t.Fatalf("ExchangeCode() error = %v", err)
	}
	if token.ExpiresIn != 7200 {
		t.Fatalf("ExpiresIn = %d, want 7200", token.ExpiresIn)
	}
	if token.CreatedAtUnix != now.Unix() {
		t.Fatalf("CreatedAtUnix = %d, want %d", token.CreatedAtUnix, now.Unix())
	}
	if token.ExpiresAtUnix != now.Add(2*time.Hour).Unix() {
		t.Fatalf("ExpiresAtUnix = %d, want %d", token.ExpiresAtUnix, now.Add(2*time.Hour).Unix())
	}
}

func TestRefreshTokenUsesRefreshGrant(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		auth := r.Header.Get("Authorization")
		wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("client-1:secret-1"))
		if auth != wantAuth {
			t.Fatalf("Authorization = %q, want %q", auth, wantAuth)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		values, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatalf("ParseQuery() error = %v", err)
		}
		if values.Get("grant_type") != "refresh_token" {
			t.Fatalf("grant_type = %q, want refresh_token", values.Get("grant_type"))
		}
		if values.Get("refresh_token") != "refresh-1" {
			t.Fatalf("refresh_token = %q, want refresh-1", values.Get("refresh_token"))
		}
		_, _ = w.Write([]byte(`{"access_token":"access-2","refresh_token":"refresh-2","token_type":"Bearer","expires_in":3600}`))
	}))
	defer server.Close()

	now := time.Unix(1_776_351_966, 0).UTC()
	token, err := Exchanger{
		HTTPClient: server.Client(),
		TokenURL:   server.URL,
		Now:        func() time.Time { return now },
	}.RefreshToken(context.Background(), OAuthConfig{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
	}, "refresh-1")
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if token.AccessToken != "access-2" {
		t.Fatalf("AccessToken = %q, want access-2", token.AccessToken)
	}
	if token.RefreshToken != "refresh-2" {
		t.Fatalf("RefreshToken = %q, want refresh-2", token.RefreshToken)
	}
	if token.ExpiresAtUnix != now.Add(time.Hour).Unix() {
		t.Fatalf("ExpiresAtUnix = %d, want %d", token.ExpiresAtUnix, now.Add(time.Hour).Unix())
	}
}
