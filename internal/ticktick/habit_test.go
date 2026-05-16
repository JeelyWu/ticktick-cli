package ticktick

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeelywu/ticktick-cli/internal/domain"
)

func TestListHabits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("Method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/open/v1/habit"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		resp := []habitDTO{
			{ID: "h1", Name: "Read", Goal: 30, Status: 0},
			{ID: "h2", Name: "Run", Goal: 5, Status: 0},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	habits, err := client.ListHabits(context.Background(), "token")
	if err != nil {
		t.Fatalf("ListHabits() error = %v", err)
	}
	if len(habits) != 2 {
		t.Fatalf("len(habits) = %d, want 2", len(habits))
	}
	// Verify sorted by name
	if habits[0].Name != "Read" {
		t.Fatalf("Name = %q, want Read", habits[0].Name)
	}
}

func TestGetHabit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/open/v1/habit/h1"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		resp := habitDTO{ID: "h1", Name: "Read", Goal: 30, Status: 0}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	habit, err := client.GetHabit(context.Background(), "token", "h1")
	if err != nil {
		t.Fatalf("GetHabit() error = %v", err)
	}
	if habit.ID != "h1" {
		t.Fatalf("ID = %q, want h1", habit.ID)
	}
}

func TestCreateHabit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("Method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/open/v1/habit"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		resp := habitDTO{ID: "h3", Name: "Meditate", Goal: 1, Status: 0}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	habit, err := client.CreateHabit(context.Background(), "token", domain.CreateHabitPayload{
		Name: "Meditate",
		Goal: 1,
	})
	if err != nil {
		t.Fatalf("CreateHabit() error = %v", err)
	}
	if habit.Name != "Meditate" {
		t.Fatalf("Name = %q, want Meditate", habit.Name)
	}
}

func TestCreateHabitWithOptionalFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode error = %v", err)
		}
		if got, want := body["color"], "red"; got != want {
			t.Fatalf("color = %v, want %v", got, want)
		}
		if got, want := body["targetDays"], float64(30); got != want {
			t.Fatalf("targetDays = %v, want %v", got, want)
		}
		resp := habitDTO{ID: "h4", Name: "Swim", Goal: 3}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	_, err := client.CreateHabit(context.Background(), "token", domain.CreateHabitPayload{
		Name:       "Swim",
		Goal:       3,
		Color:      "red",
		TargetDays: 30,
	})
	if err != nil {
		t.Fatalf("CreateHabit() error = %v", err)
	}
}

func TestUpdateHabit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/open/v1/habit/h1"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode error = %v", err)
		}
		if got, want := body["id"], "h1"; got != want {
			t.Fatalf("id = %v, want %v", got, want)
		}
		if _, ok := body["status"]; ok {
			t.Fatal("status should not be present when StatusSet is false")
		}
		resp := habitDTO{ID: "h1", Name: "Read More", Goal: 50}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	habit, err := client.UpdateHabit(context.Background(), "token", "h1", domain.CreateHabitPayload{
		Name: "Read More",
		Goal: 50,
	})
	if err != nil {
		t.Fatalf("UpdateHabit() error = %v", err)
	}
	if habit.Goal != 50 {
		t.Fatalf("Goal = %v, want 50", habit.Goal)
	}
}

func TestUpdateHabitWithStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode error = %v", err)
		}
		if got, want := body["status"], float64(1); got != want {
			t.Fatalf("status = %v, want %v", got, want)
		}
		resp := habitDTO{ID: "h1", Name: "Read", Goal: 30, Status: 1}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	_, err := client.UpdateHabit(context.Background(), "token", "h1", domain.CreateHabitPayload{
		Name:      "Read",
		Goal:      30,
		Status:    1,
		StatusSet: true,
	})
	if err != nil {
		t.Fatalf("UpdateHabit() error = %v", err)
	}
}

func TestCheckinHabit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/open/v1/habit/h1/checkin"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode error = %v", err)
		}
		if got, want := body["stamp"], float64(20260513); got != want {
			t.Fatalf("stamp = %v, want %v", got, want)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	if err := client.CheckinHabit(context.Background(), "token", "h1", 20260513, 1, 1); err != nil {
		t.Fatalf("CheckinHabit() error = %v", err)
	}
}

func TestCheckinHabitWithoutGoal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode error = %v", err)
		}
		if _, ok := body["goal"]; ok {
			t.Fatal("goal should not be present when zero")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	if err := client.CheckinHabit(context.Background(), "token", "h1", 20260513, 1, 0); err != nil {
		t.Fatalf("CheckinHabit() error = %v", err)
	}
}

func TestListCheckins(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("Method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/open/v1/habit/checkins"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		q := r.URL.Query()
		if got, want := q.Get("habitIds"), "h1,h2"; got != want {
			t.Fatalf("habitIds = %q, want %q", got, want)
		}
		resp := []habitCheckinGroupDTO{
			{HabitID: "h1", Year: 2026, Checkins: []checkinEntryDTO{{Stamp: 20260513, Value: 1, Goal: 1}}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	checkins, err := client.ListCheckins(context.Background(), "token", []string{"h1", "h2"}, 20260101, 20261231)
	if err != nil {
		t.Fatalf("ListCheckins() error = %v", err)
	}
	if len(checkins) != 1 {
		t.Fatalf("len(checkins) = %d, want 1", len(checkins))
	}
	if checkins[0].Stamp != 20260513 {
		t.Fatalf("Stamp = %d, want 20260513", checkins[0].Stamp)
	}
}

func TestListHabitsReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	_, err := client.ListHabits(context.Background(), "token")
	if err == nil {
		t.Fatal("ListHabits() error = nil, want error")
	}
}

func TestMapHabits(t *testing.T) {
	dtos := []habitDTO{
		{ID: "h1", Name: "Beta", Goal: 1, Status: 0},
		{ID: "h2", Name: "Alpha", Goal: 2, Status: 1},
	}
	habits := mapHabits(dtos)
	if len(habits) != 2 {
		t.Fatalf("len(habits) = %d, want 2", len(habits))
	}
	// Verify sorted by name
	if habits[0].Name != "Alpha" {
		t.Fatalf("Name = %q, want Alpha", habits[0].Name)
	}
	if habits[0].Status != domain.HabitStatusArchived {
		t.Fatalf("Status = %v, want archived", habits[0].Status)
	}
	if habits[1].Status != domain.HabitStatusActive {
		t.Fatalf("Status = %v, want active", habits[1].Status)
	}
}

func TestMapCheckins(t *testing.T) {
	group := habitCheckinGroupDTO{
		HabitID: "h1",
		Year:    2026,
		Checkins: []checkinEntryDTO{
			{Stamp: 20260510, Value: 1, Goal: 1},
			{Stamp: 20260513, Value: 1, Goal: 1},
		},
	}
	checkins := mapCheckins(group)
	if len(checkins) != 2 {
		t.Fatalf("len(checkins) = %d, want 2", len(checkins))
	}
	// Verify sorted by stamp descending
	if checkins[0].Stamp != 20260513 {
		t.Fatalf("Stamp = %d, want 20260513", checkins[0].Stamp)
	}
}
