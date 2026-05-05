package app

import (
	"context"
	"testing"
	"time"

	"github.com/jeely/ticktick-cli/internal/domain"
)

type recordingFocusAPI struct {
	projects      []domain.Project
	focuses       []domain.Focus
	lastListStart time.Time
	lastListEnd   time.Time
	startCalls    []domain.StartFocusInput
	stopCalls     []string
}

func (r *recordingFocusAPI) ListProjects(context.Context, string) ([]domain.Project, error) {
	if r.projects != nil {
		return r.projects, nil
	}
	return []domain.Project{{ID: "p1", Name: "Inbox"}}, nil
}

func (r *recordingFocusAPI) GetFocus(_ context.Context, _ string, focusID string) (domain.Focus, error) {
	for _, f := range r.focuses {
		if f.ID == focusID {
			return f, nil
		}
	}
	return domain.Focus{}, nil
}

func (r *recordingFocusAPI) ListFocus(_ context.Context, _ string, startDate, endDate time.Time) ([]domain.Focus, error) {
	r.lastListStart = startDate
	r.lastListEnd = endDate
	return r.focuses, nil
}

func (r *recordingFocusAPI) StartFocus(_ context.Context, _ string, in domain.StartFocusInput) (domain.Focus, error) {
	r.startCalls = append(r.startCalls, in)
	return domain.Focus{ID: "f1", Title: in.Title, Mode: in.Mode, ProjectID: in.ProjectID}, nil
}

func (r *recordingFocusAPI) StopFocus(_ context.Context, _ string, focusID string) error {
	r.stopCalls = append(r.stopCalls, focusID)
	return nil
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

	focuses, _, err := focusApp.List(context.Background(), ListFocusInput{})
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
}

func TestFocusAppListFiltersByProject(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.Local)
	client := &recordingFocusAPI{
		projects: []domain.Project{{ID: "p1", Name: "Inbox"}, {ID: "p2", Name: "Work"}},
		focuses: []domain.Focus{
			{ID: "f1", Title: "Personal", ProjectID: "p1"},
			{ID: "f2", Title: "Work", ProjectID: "p2"},
		},
	}
	focusApp := FocusApp{
		Auth:   stubTokenSource{},
		Client: client,
		Now: func() time.Time {
			return now
		},
	}

	focuses, _, err := focusApp.List(context.Background(), ListFocusInput{Project: "Work"})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(focuses) != 1 {
		t.Fatalf("len(focuses) = %d, want 1", len(focuses))
	}
	if focuses[0].ID != "f2" {
		t.Fatalf("focuses[0].ID = %q, want f2", focuses[0].ID)
	}
}

func TestFocusAppStartResolvesProject(t *testing.T) {
	client := &recordingFocusAPI{
		projects: []domain.Project{{ID: "p2", Name: "Work"}},
	}
	focusApp := FocusApp{
		Auth:   stubTokenSource{},
		Client: client,
	}

	focus, err := focusApp.Start(context.Background(), StartFocusAppInput{
		Title:      "Deep work",
		ProjectRef: "Work",
		Mode:       domain.FocusModeTimer,
	})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if focus.ProjectID != "p2" {
		t.Fatalf("focus.ProjectID = %q, want p2", focus.ProjectID)
	}
	if len(client.startCalls) != 1 {
		t.Fatalf("start calls = %d, want 1", len(client.startCalls))
	}
	if client.startCalls[0].ProjectID != "p2" {
		t.Fatalf("start call ProjectID = %q, want p2", client.startCalls[0].ProjectID)
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

	focus, err := focusApp.Get(context.Background(), "f1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if focus.ID != "f1" {
		t.Fatalf("focus.ID = %q, want f1", focus.ID)
	}
	if focus.Title != "Deep work" {
		t.Fatalf("focus.Title = %q, want Deep work", focus.Title)
	}
}

func TestFocusAppStartWithCustomTime(t *testing.T) {
	client := &recordingFocusAPI{
		projects: []domain.Project{{ID: "p2", Name: "Work"}},
	}
	focusApp := FocusApp{
		Auth:   stubTokenSource{},
		Client: client,
	}

	_, err := focusApp.Start(context.Background(), StartFocusAppInput{
		Title:      "Deep work",
		ProjectRef: "Work",
		Mode:       domain.FocusModeTimer,
		StartRaw:   "2026-05-01",
	})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if len(client.startCalls) != 1 {
		t.Fatalf("start calls = %d, want 1", len(client.startCalls))
	}
	if client.startCalls[0].StartDate == nil {
		t.Fatalf("start call StartDate = nil, want non-nil")
	}
	want := time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local)
	if !client.startCalls[0].StartDate.Equal(want) {
		t.Fatalf("start call StartDate = %v, want %v", client.startCalls[0].StartDate, want)
	}
}

func TestFocusAppStop(t *testing.T) {
	client := &recordingFocusAPI{}
	focusApp := FocusApp{
		Auth:   stubTokenSource{},
		Client: client,
	}

	if err := focusApp.Stop(context.Background(), "f1"); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if len(client.stopCalls) != 1 || client.stopCalls[0] != "f1" {
		t.Fatalf("stop calls = %v, want [f1]", client.stopCalls)
	}
}
