package ticktick

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
