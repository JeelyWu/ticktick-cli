package ticktick

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jeely/ticktick-cli/internal/domain"
)

func TestGetFocus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("Method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/open/v1/focus/f1"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("type"), "1"; got != want {
			t.Fatalf("type = %q, want %q", got, want)
		}
		resp := focusDTO{ID: "f1", Note: "Focus 1", Type: 1, Status: 0}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	focus, err := client.GetFocus(context.Background(), "token", "f1", 1)
	if err != nil {
		t.Fatalf("GetFocus() error = %v", err)
	}
	if focus.ID != "f1" {
		t.Fatalf("ID = %q, want f1", focus.ID)
	}
	if focus.Title != "Focus 1" {
		t.Fatalf("Title = %q, want Focus 1", focus.Title)
	}
	if focus.Mode != domain.FocusModeTimer {
		t.Fatalf("Mode = %v, want timer", focus.Mode)
	}
}

func TestListFocus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("Method = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("type"), "0"; got != want {
			t.Fatalf("type = %q, want %q", got, want)
		}
		resp := focusListResponse{Focuses: []focusDTO{
			{ID: "f1", Note: "Focus 1", Type: 0, Status: 0},
		}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	start := time.Now()
	end := start.Add(24 * time.Hour)
	focuses, err := client.ListFocus(context.Background(), "token", start, end, 0)
	if err != nil {
		t.Fatalf("ListFocus() error = %v", err)
	}
	if len(focuses) != 1 {
		t.Fatalf("len(focuses) = %d, want 1", len(focuses))
	}
	if focuses[0].Mode != domain.FocusModePomodoro {
		t.Fatalf("Mode = %v, want pomodoro", focuses[0].Mode)
	}
}

func TestListFocusAcceptsArrayResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []focusDTO{
			{ID: "f1", Note: "Focus 1", Type: 1, Status: 1},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	start := time.Now()
	end := start.Add(24 * time.Hour)
	focuses, err := client.ListFocus(context.Background(), "token", start, end, 1)
	if err != nil {
		t.Fatalf("ListFocus() error = %v", err)
	}
	if len(focuses) != 1 {
		t.Fatalf("len(focuses) = %d, want 1", len(focuses))
	}
	if focuses[0].Mode != domain.FocusModeTimer {
		t.Fatalf("Mode = %v, want timer", focuses[0].Mode)
	}
}

func TestGetFocusReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	_, err := client.GetFocus(context.Background(), "token", "f1", 1)
	if err == nil {
		t.Fatal("GetFocus() error = nil, want error")
	}
}

func TestMapFocus(t *testing.T) {
	now := time.Now()
	dto := focusDTO{
		ID:        "f1",
		Type:      0,
		Status:    0,
		Note:      "Focus",
		TaskID:    "t1",
		StartTime: now.Format("2006-01-02T15:04:05.000-0700"),
	}
	f := mapFocus(dto)
	if f.ID != "f1" {
		t.Fatalf("ID = %q, want f1", f.ID)
	}
	if f.Mode != domain.FocusModePomodoro {
		t.Fatalf("Mode = %v, want pomodoro", f.Mode)
	}
	if f.Status != domain.FocusStatusActive {
		t.Fatalf("Status = %v, want active", f.Status)
	}
	if f.Title != "Focus" {
		t.Fatalf("Title = %q, want Focus", f.Title)
	}
	if f.StartDate == nil {
		t.Fatal("StartDate = nil, want non-nil")
	}
}
