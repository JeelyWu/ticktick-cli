package app

import (
	"context"
	"testing"
	"time"

	"github.com/jeelywu/ticktick-cli/internal/domain"
)

type recordingFocusAPI struct {
	projects      []domain.Project
	focuses       []domain.Focus
	lastListStart time.Time
	lastListEnd   time.Time
	lastListType  int
	lastGetType   int
}

func (r *recordingFocusAPI) ListProjects(context.Context, string) ([]domain.Project, error) {
	if r.projects != nil {
		return r.projects, nil
	}
	return []domain.Project{{ID: "p1", Name: "Inbox"}}, nil
}

func (r *recordingFocusAPI) GetFocus(_ context.Context, _ string, focusID string, focusType int) (domain.Focus, error) {
	r.lastGetType = focusType
	for _, f := range r.focuses {
		if f.ID == focusID {
			return f, nil
		}
	}
	return domain.Focus{}, nil
}

func (r *recordingFocusAPI) ListFocus(_ context.Context, _ string, startDate, endDate time.Time, focusType int) ([]domain.Focus, error) {
	r.lastListStart = startDate
	r.lastListEnd = endDate
	r.lastListType = focusType
	return r.focuses, nil
}

func TestFocusAppListDefaultTimeRange(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.Local)
	client := &recordingFocusAPI{focuses: []domain.Focus{{ID: "f1", Title: "Deep work", ProjectID: "p1"}}}
	focusApp := FocusApp{
		Auth:   stubTokenSource{},
		Client: client,
		Now: func() time.Time {
			return now
		},
	}

	focuses, _, err := focusApp.List(context.Background(), ListFocusInput{Type: 1})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(focuses) != 1 {
		t.Fatalf("len(focuses) = %d, want 1", len(focuses))
	}

	wantStart := now.AddDate(0, 0, -7)
	if !client.lastListStart.Equal(wantStart) {
		t.Fatalf("list start = %v, want %v", client.lastListStart, wantStart)
	}
	if !client.lastListEnd.Equal(now) {
		t.Fatalf("list end = %v, want %v", client.lastListEnd, now)
	}
	if client.lastListType != 1 {
		t.Fatalf("list type = %d, want 1", client.lastListType)
	}
}

func TestFocusAppListCustomTimeRange(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.Local)
	client := &recordingFocusAPI{focuses: []domain.Focus{{ID: "f1", Title: "Deep work", ProjectID: "p1"}}}
	focusApp := FocusApp{
		Auth:   stubTokenSource{},
		Client: client,
		Now: func() time.Time {
			return now
		},
	}

	_, _, err := focusApp.List(context.Background(), ListFocusInput{
		From: "2026-05-01",
		To:   "2026-05-05",
		Type: 0,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	wantStart := time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local)
	wantEnd := time.Date(2026, 5, 5, 0, 0, 0, 0, time.Local)
	if !client.lastListStart.Equal(wantStart) {
		t.Fatalf("list start = %v, want %v", client.lastListStart, wantStart)
	}
	if !client.lastListEnd.Equal(wantEnd) {
		t.Fatalf("list end = %v, want %v", client.lastListEnd, wantEnd)
	}
	if client.lastListType != 0 {
		t.Fatalf("list type = %d, want 0", client.lastListType)
	}
}

func TestFocusAppGet(t *testing.T) {
	client := &recordingFocusAPI{
		focuses: []domain.Focus{{ID: "f1", Title: "Deep work"}},
	}
	focusApp := FocusApp{
		Auth:   stubTokenSource{},
		Client: client,
	}

	focus, err := focusApp.Get(context.Background(), "f1", 0)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if focus.ID != "f1" {
		t.Fatalf("focus.ID = %q, want f1", focus.ID)
	}
	if focus.Title != "Deep work" {
		t.Fatalf("focus.Title = %q, want Deep work", focus.Title)
	}
	if client.lastGetType != 0 {
		t.Fatalf("get type = %d, want 0", client.lastGetType)
	}
}
