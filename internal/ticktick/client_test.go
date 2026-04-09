package ticktick

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewDefaultsBaseURLAndTimeout(t *testing.T) {
	client := New("", nil)

	if got, want := client.BaseURL, "https://api.ticktick.com"; got != want {
		t.Fatalf("BaseURL = %q, want %q", got, want)
	}
	if client.HTTPClient == nil {
		t.Fatal("HTTPClient = nil, want default client")
	}
	if client.HTTPClient.Timeout == 0 {
		t.Fatal("HTTPClient.Timeout = 0, want a sane default timeout")
	}
}

func TestNewPreservesInjectedHTTPClient(t *testing.T) {
	injected := &http.Client{Timeout: 5 * time.Second}

	client := New("https://example.com/", injected)

	if client.HTTPClient != injected {
		t.Fatal("HTTPClient was replaced, want injected client preserved")
	}
	if got, want := client.BaseURL, "https://example.com"; got != want {
		t.Fatalf("BaseURL = %q, want %q", got, want)
	}
}

func TestDoJSONAddsBearerToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("Authorization header = %q, want Bearer token-123", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	var out struct {
		OK bool `json:"ok"`
	}
	if err := client.DoJSON(context.Background(), http.MethodGet, "/ping", "token-123", nil, &out); err != nil {
		t.Fatalf("DoJSON() error = %v", err)
	}
	if !out.OK {
		t.Fatalf("out.OK = false, want true")
	}
}

func TestDoJSONReturnsRemoteError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	err := client.DoJSON(context.Background(), http.MethodGet, "/ping", "bad-token", nil, nil)
	if err == nil {
		t.Fatal("DoJSON() error = nil, want remote error")
	}
	if _, ok := err.(*RemoteError); !ok {
		t.Fatalf("error type = %T, want *RemoteError", err)
	}
}

func TestDoJSONEncodesRequestJSONAndTreatsEmptySuccessAsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
			t.Fatalf("Content-Type = %q, want %q", got, want)
		}
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("Method = %q, want %q", got, want)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if got, want := string(body), `{"name":"alpha","count":2}`; got != want {
			t.Fatalf("request body = %q, want %q", got, want)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	var out struct {
		OK bool `json:"ok"`
	}
	err := client.DoJSON(context.Background(), http.MethodPost, "/items", "token-123", struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}{Name: "alpha", Count: 2}, &out)
	if err != nil {
		t.Fatalf("DoJSON() error = %v", err)
	}
	if out != (struct {
		OK bool `json:"ok"`
	}{}) {
		t.Fatalf("out = %#v, want zero value after empty success response", out)
	}
}
