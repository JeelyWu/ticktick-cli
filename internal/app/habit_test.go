package app

import (
	"context"
	"errors"
	"testing"

	"github.com/jeelywu/ticktick-cli/internal/domain"
)

type stubHabitAPI struct {
	listHabits   func(context.Context, string) ([]domain.Habit, error)
	getHabit     func(context.Context, string, string) (domain.Habit, error)
	createHabit  func(context.Context, string, domain.CreateHabitPayload) (domain.Habit, error)
	updateHabit  func(context.Context, string, string, domain.CreateHabitPayload) (domain.Habit, error)
	checkinHabit func(context.Context, string, string, int, float64, float64) error
	listCheckins func(context.Context, string, []string, int, int) ([]domain.HabitCheckin, error)
}

func (s *stubHabitAPI) ListHabits(ctx context.Context, token string) ([]domain.Habit, error) {
	if s.listHabits != nil {
		return s.listHabits(ctx, token)
	}
	return nil, nil
}

func (s *stubHabitAPI) GetHabit(ctx context.Context, token, id string) (domain.Habit, error) {
	if s.getHabit != nil {
		return s.getHabit(ctx, token, id)
	}
	return domain.Habit{}, nil
}

func (s *stubHabitAPI) CreateHabit(ctx context.Context, token string, in domain.CreateHabitPayload) (domain.Habit, error) {
	if s.createHabit != nil {
		return s.createHabit(ctx, token, in)
	}
	return domain.Habit{}, nil
}

func (s *stubHabitAPI) UpdateHabit(ctx context.Context, token, id string, in domain.CreateHabitPayload) (domain.Habit, error) {
	if s.updateHabit != nil {
		return s.updateHabit(ctx, token, id, in)
	}
	return domain.Habit{}, nil
}

func (s *stubHabitAPI) CheckinHabit(ctx context.Context, token, id string, stamp int, value, goal float64) error {
	if s.checkinHabit != nil {
		return s.checkinHabit(ctx, token, id, stamp, value, goal)
	}
	return nil
}

func (s *stubHabitAPI) ListCheckins(ctx context.Context, token string, ids []string, from, to int) ([]domain.HabitCheckin, error) {
	if s.listCheckins != nil {
		return s.listCheckins(ctx, token, ids, from, to)
	}
	return nil, nil
}

type fakeTokenSource struct {
	token string
	err   error
}

func (f *fakeTokenSource) AccessToken(_ context.Context) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.token, nil
}

func newTestHabitApp(client HabitAPI) *HabitApp {
	return &HabitApp{
		Auth:   &fakeTokenSource{token: "tok"},
		Client: client,
	}
}

func TestHabitList(t *testing.T) {
	client := &stubHabitAPI{
		listHabits: func(_ context.Context, _ string) ([]domain.Habit, error) {
			return []domain.Habit{
				{ID: "h1", Name: "Exercise"},
				{ID: "h2", Name: "Read"},
			}, nil
		},
	}
	app := newTestHabitApp(client)
	habits, err := app.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(habits) != 2 {
		t.Fatalf("len(habits) = %d, want 2", len(habits))
	}
	if habits[0].Name != "Exercise" {
		t.Fatalf("habits[0].Name = %q, want Exercise", habits[0].Name)
	}
}

func TestHabitGetByName(t *testing.T) {
	client := &stubHabitAPI{
		listHabits: func(_ context.Context, _ string) ([]domain.Habit, error) {
			return []domain.Habit{{ID: "h1", Name: "Exercise"}}, nil
		},
	}
	app := newTestHabitApp(client)
	habit, err := app.Get(context.Background(), "Exercise")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if habit.ID != "h1" {
		t.Fatalf("habit.ID = %q, want h1", habit.ID)
	}
}

func TestHabitGetByID(t *testing.T) {
	client := &stubHabitAPI{
		getHabit: func(_ context.Context, _, id string) (domain.Habit, error) {
			if id == "abc123def456789012345678" {
				return domain.Habit{ID: "abc123def456789012345678", Name: "Exercise"}, nil
			}
			return domain.Habit{}, errors.New("not found")
		},
	}
	app := newTestHabitApp(client)
	habit, err := app.Get(context.Background(), "abc123def456789012345678")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if habit.ID != "abc123def456789012345678" {
		t.Fatalf("habit.ID = %q, want abc123def456789012345678", habit.ID)
	}
}

func TestHabitGetFallbackToAPI(t *testing.T) {
	client := &stubHabitAPI{
		listHabits: func(_ context.Context, _ string) ([]domain.Habit, error) {
			return []domain.Habit{{ID: "h1", Name: "Other"}}, nil
		},
		getHabit: func(_ context.Context, _, id string) (domain.Habit, error) {
			if id == "archived-habit" {
				return domain.Habit{ID: "archived-habit", Name: "Archived Habit", Status: domain.HabitStatusArchived}, nil
			}
			return domain.Habit{}, errors.New("not found")
		},
	}
	app := newTestHabitApp(client)
	habit, err := app.Get(context.Background(), "archived-habit")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if habit.ID != "archived-habit" {
		t.Fatalf("habit.ID = %q, want archived-habit", habit.ID)
	}
	if habit.Status != domain.HabitStatusArchived {
		t.Fatalf("habit.Status = %v, want archived", habit.Status)
	}
}

func TestHabitGetNotFound(t *testing.T) {
	client := &stubHabitAPI{
		listHabits: func(_ context.Context, _ string) ([]domain.Habit, error) {
			return []domain.Habit{}, nil
		},
		getHabit: func(_ context.Context, _, _ string) (domain.Habit, error) {
			return domain.Habit{}, errors.New("not found")
		},
	}
	app := newTestHabitApp(client)
	_, err := app.Get(context.Background(), "Missing")
	if err == nil {
		t.Fatal("Get() error = nil, want error")
	}
	var refErr *domain.ReferenceError
	if !errors.As(err, &refErr) {
		t.Fatalf("error type = %T, want *domain.ReferenceError", err)
	}
}

func TestHabitAddDefaults(t *testing.T) {
	var captured domain.CreateHabitPayload
	client := &stubHabitAPI{
		createHabit: func(_ context.Context, _ string, in domain.CreateHabitPayload) (domain.Habit, error) {
			captured = in
			return domain.Habit{ID: "h1", Name: in.Name, Goal: in.Goal, Step: in.Step}, nil
		},
	}
	app := newTestHabitApp(client)
	habit, err := app.Add(context.Background(), domain.CreateHabitInput{Name: "Run"})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if captured.Goal != 1 {
		t.Fatalf("captured.Goal = %v, want 1", captured.Goal)
	}
	if captured.Step != 1 {
		t.Fatalf("captured.Step = %v, want 1", captured.Step)
	}
	if habit.Name != "Run" {
		t.Fatalf("habit.Name = %q, want Run", habit.Name)
	}
}

func TestHabitArchiveToggle(t *testing.T) {
	var captured domain.CreateHabitPayload
	client := &stubHabitAPI{
		listHabits: func(_ context.Context, _ string) ([]domain.Habit, error) {
			return []domain.Habit{{ID: "h1", Name: "Run", Status: domain.HabitStatusActive}}, nil
		},
		updateHabit: func(_ context.Context, _ string, _ string, in domain.CreateHabitPayload) (domain.Habit, error) {
			captured = in
			return domain.Habit{ID: "h1", Name: "Run", Status: domain.HabitStatus(captured.Status)}, nil
		},
	}
	app := newTestHabitApp(client)
	habit, err := app.Archive(context.Background(), "Run")
	if err != nil {
		t.Fatalf("Archive() error = %v", err)
	}
	if !captured.StatusSet {
		t.Fatal("Archive() did not set StatusSet")
	}
	if captured.Status != 1 {
		t.Fatalf("captured.Status = %d, want 1", captured.Status)
	}
	if habit.Status != domain.HabitStatusArchived {
		t.Fatalf("habit.Status = %v, want archived", habit.Status)
	}
}

func TestHabitCheckin(t *testing.T) {
	var capturedStamp int
	var capturedValue float64
	client := &stubHabitAPI{
		listHabits: func(_ context.Context, _ string) ([]domain.Habit, error) {
			return []domain.Habit{{ID: "h1", Name: "Run", Step: 1, Goal: 1}}, nil
		},
		checkinHabit: func(_ context.Context, _ string, _ string, stamp int, value, goal float64) error {
			capturedStamp = stamp
			capturedValue = value
			return nil
		},
	}
	app := newTestHabitApp(client)
	if err := app.Checkin(context.Background(), "Run", 0); err != nil {
		t.Fatalf("Checkin() error = %v", err)
	}
	if capturedStamp == 0 {
		t.Fatal("Checkin() did not set stamp")
	}
	if capturedValue != 1 {
		t.Fatalf("capturedValue = %v, want 1", capturedValue)
	}
}

func TestResolveHabit(t *testing.T) {
	habits := []domain.Habit{
		{ID: "h1", Name: "Exercise"},
		{ID: "h2", Name: "Read"},
	}

	habit, err := ResolveHabit("Read", habits)
	if err != nil {
		t.Fatalf("ResolveHabit() error = %v", err)
	}
	if habit.ID != "h2" {
		t.Fatalf("habit.ID = %q, want h2", habit.ID)
	}

	habit, err = ResolveHabit("h1", habits)
	if err != nil {
		t.Fatalf("ResolveHabit() by id error = %v", err)
	}
	if habit.Name != "Exercise" {
		t.Fatalf("habit.Name = %q, want Exercise", habit.Name)
	}
}

func TestResolveHabitAmbiguous(t *testing.T) {
	habits := []domain.Habit{
		{ID: "h1", Name: "Exercise"},
		{ID: "h2", Name: "Exercise"},
	}
	if _, err := ResolveHabit("Exercise", habits); err == nil {
		t.Fatal("ResolveHabit() error = nil, want ambiguity error")
	}
}
