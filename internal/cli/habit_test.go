package cli

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jeelywu/ticktick-cli/internal/app"
	"github.com/jeelywu/ticktick-cli/internal/domain"
)

func TestHabitListJSON(t *testing.T) {
	streams, stdout, _ := newTestStreams()
	cmd := NewHabitCommand(func() (*app.HabitApp, error) {
		return &app.HabitApp{
			Auth: fakeTokenSource{},
			Client: &stubHabitClient{
				listHabits: func() ([]domain.Habit, error) {
					return []domain.Habit{{ID: "h1", Name: "Run", Goal: 1, Status: domain.HabitStatusActive}}, nil
				},
			},
		}, nil
	}, nil, streams)
	cmd.SetArgs([]string{"ls", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), `"id": "h1"`) {
		t.Fatalf("output = %q, want h1", stdout.String())
	}
}

func TestHabitGetNotFound(t *testing.T) {
	streams, _, _ := newTestStreams()
	cmd := NewHabitCommand(func() (*app.HabitApp, error) {
		return &app.HabitApp{
			Auth: fakeTokenSource{},
			Client: &stubHabitClient{
				listHabits: func() ([]domain.Habit, error) {
					return []domain.Habit{}, nil
				},
			},
		}, nil
	}, nil, streams)
	cmd.SetArgs([]string{"get", "Missing"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "Missing") {
		t.Fatalf("error = %q, want Missing", err.Error())
	}
}

func TestHabitAddDefaults(t *testing.T) {
	streams, _, _ := newTestStreams()
	var captured domain.CreateHabitPayload
	cmd := NewHabitCommand(func() (*app.HabitApp, error) {
		return &app.HabitApp{
			Auth: fakeTokenSource{},
			Client: &stubHabitClient{
				createHabit: func(in domain.CreateHabitPayload) (domain.Habit, error) {
					captured = in
					return domain.Habit{ID: "h1", Name: in.Name}, nil
				},
			},
		}, nil
	}, nil, streams)
	cmd.SetArgs([]string{"add", "Run"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if captured.Name != "Run" {
		t.Fatalf("captured.Name = %q, want Run", captured.Name)
	}
	if captured.Goal != 1 {
		t.Fatalf("captured.Goal = %v, want 1", captured.Goal)
	}
	if captured.Step != 1 {
		t.Fatalf("captured.Step = %v, want 1", captured.Step)
	}
}

func TestHabitArchive(t *testing.T) {
	streams, stdout, _ := newTestStreams()
	cmd := NewHabitCommand(func() (*app.HabitApp, error) {
		return &app.HabitApp{
			Auth: fakeTokenSource{},
			Client: &stubHabitClient{
				listHabits: func() ([]domain.Habit, error) {
					return []domain.Habit{{ID: "h1", Name: "Run", Status: domain.HabitStatusActive}}, nil
				},
				updateHabit: func(id string, in domain.CreateHabitPayload) (domain.Habit, error) {
					return domain.Habit{ID: "h1", Name: "Run", Status: domain.HabitStatusArchived}, nil
				},
			},
		}, nil
	}, nil, streams)
	cmd.SetArgs([]string{"archive", "Run"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "archived") {
		t.Fatalf("output = %q, want archived", stdout.String())
	}
}

func TestHabitCheckin(t *testing.T) {
	streams, stdout, _ := newTestStreams()
	cmd := NewHabitCommand(func() (*app.HabitApp, error) {
		return &app.HabitApp{
			Auth: fakeTokenSource{},
			Client: &stubHabitClient{
				listHabits: func() ([]domain.Habit, error) {
					return []domain.Habit{{ID: "h1", Name: "Run", Step: 1, Goal: 1}}, nil
				},
				checkinHabit: func(id string, stamp int, value, goal float64) error {
					return nil
				},
			},
		}, nil
	}, nil, streams)
	cmd.SetArgs([]string{"checkin", "Run"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Checked in") {
		t.Fatalf("output = %q, want Checked in", stdout.String())
	}
}

func TestHabitCheckinWithValue(t *testing.T) {
	streams, _, _ := newTestStreams()
	var capturedValue float64
	cmd := NewHabitCommand(func() (*app.HabitApp, error) {
		return &app.HabitApp{
			Auth: fakeTokenSource{},
			Client: &stubHabitClient{
				listHabits: func() ([]domain.Habit, error) {
					return []domain.Habit{{ID: "h1", Name: "Run", Step: 1, Goal: 1}}, nil
				},
				checkinHabit: func(id string, stamp int, value, goal float64) error {
					capturedValue = value
					return nil
				},
			},
		}, nil
	}, nil, streams)
	cmd.SetArgs([]string{"checkin", "Run", "--value", "3"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if capturedValue != 3 {
		t.Fatalf("capturedValue = %v, want 3", capturedValue)
	}
}

func TestHabitLog(t *testing.T) {
	streams, stdout, _ := newTestStreams()
	cmd := NewHabitCommand(func() (*app.HabitApp, error) {
		return &app.HabitApp{
			Auth: fakeTokenSource{},
			Client: &stubHabitClient{
				listHabits: func() ([]domain.Habit, error) {
					return []domain.Habit{{ID: "h1", Name: "Run"}}, nil
				},
				listCheckins: func(ids []string, from, to int) ([]domain.HabitCheckin, error) {
					return []domain.HabitCheckin{
						{HabitID: "h1", Stamp: 20260506, Value: 1, Goal: 1},
					}, nil
				},
			},
		}, nil
	}, nil, streams)
	cmd.SetArgs([]string{"log", "Run"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "2026-05-06") {
		t.Fatalf("output = %q, want date", stdout.String())
	}
}

func TestHabitUpdateStatus(t *testing.T) {
	streams, _, _ := newTestStreams()
	var captured domain.CreateHabitPayload
	cmd := NewHabitCommand(func() (*app.HabitApp, error) {
		return &app.HabitApp{
			Auth: fakeTokenSource{},
			Client: &stubHabitClient{
				listHabits: func() ([]domain.Habit, error) {
					return []domain.Habit{{ID: "h1", Name: "Run", Goal: 1, Step: 1}}, nil
				},
				updateHabit: func(id string, in domain.CreateHabitPayload) (domain.Habit, error) {
					captured = in
					return domain.Habit{ID: "h1", Name: "Run"}, nil
				},
			},
		}, nil
	}, nil, streams)
	cmd.SetArgs([]string{"update", "Run", "--status", "archived"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !captured.StatusSet {
		t.Fatal("update did not set StatusSet")
	}
	if captured.Status != 1 {
		t.Fatalf("captured.Status = %d, want 1", captured.Status)
	}
}

func TestHabitUpdateInvalidStatus(t *testing.T) {
	streams, _, _ := newTestStreams()
	cmd := NewHabitCommand(func() (*app.HabitApp, error) {
		return &app.HabitApp{
			Auth: fakeTokenSource{},
			Client: &stubHabitClient{},
		}, nil
	}, nil, streams)
	cmd.SetArgs([]string{"update", "Run", "--status", "invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Fatalf("error = %q, want invalid status", err.Error())
	}
}

type stubHabitClient struct {
	listHabits   func() ([]domain.Habit, error)
	getHabit     func(string) (domain.Habit, error)
	createHabit  func(domain.CreateHabitPayload) (domain.Habit, error)
	updateHabit  func(string, domain.CreateHabitPayload) (domain.Habit, error)
	checkinHabit func(string, int, float64, float64) error
	listCheckins func([]string, int, int) ([]domain.HabitCheckin, error)
}

func (s *stubHabitClient) ListHabits(_ context.Context, _ string) ([]domain.Habit, error) {
	if s.listHabits != nil {
		return s.listHabits()
	}
	return nil, errors.New("ListHabits not stubbed")
}

func (s *stubHabitClient) GetHabit(_ context.Context, _, _ string) (domain.Habit, error) {
	if s.getHabit != nil {
		return s.getHabit("")
	}
	return domain.Habit{}, errors.New("GetHabit not stubbed")
}

func (s *stubHabitClient) CreateHabit(_ context.Context, _ string, in domain.CreateHabitPayload) (domain.Habit, error) {
	if s.createHabit != nil {
		return s.createHabit(in)
	}
	return domain.Habit{}, errors.New("CreateHabit not stubbed")
}

func (s *stubHabitClient) UpdateHabit(_ context.Context, _, id string, in domain.CreateHabitPayload) (domain.Habit, error) {
	if s.updateHabit != nil {
		return s.updateHabit(id, in)
	}
	return domain.Habit{}, errors.New("UpdateHabit not stubbed")
}

func (s *stubHabitClient) CheckinHabit(_ context.Context, _, id string, stamp int, value, goal float64) error {
	if s.checkinHabit != nil {
		return s.checkinHabit(id, stamp, value, goal)
	}
	return errors.New("CheckinHabit not stubbed")
}

func (s *stubHabitClient) ListCheckins(_ context.Context, _ string, ids []string, from, to int) ([]domain.HabitCheckin, error) {
	if s.listCheckins != nil {
		return s.listCheckins(ids, from, to)
	}
	return nil, errors.New("ListCheckins not stubbed")
}

var _ app.HabitAPI = (*stubHabitClient)(nil)
