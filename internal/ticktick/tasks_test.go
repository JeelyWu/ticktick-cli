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

func TestFilterTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("Method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/open/v1/task/filter"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		resp := []taskDTO{
			{ID: "t1", Title: "Task 1", ProjectID: "p1", Priority: 5, Status: 0},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	tasks, err := client.FilterTasks(context.Background(), "token", domain.TaskFilter{
		ProjectIDs: []string{"p1"},
		Priorities: []domain.Priority{domain.PriorityHigh},
		Statuses:   []domain.TaskStatus{domain.StatusOpen},
	})
	if err != nil {
		t.Fatalf("FilterTasks() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	if tasks[0].Title != "Task 1" {
		t.Fatalf("Title = %q, want Task 1", tasks[0].Title)
	}
}

func TestFilterTasksWithDates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body taskFilterRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode error = %v", err)
		}
		if body.StartDate == "" {
			t.Fatal("StartDate empty, want non-empty")
		}
		if body.EndDate == "" {
			t.Fatal("EndDate empty, want non-empty")
		}
		_ = json.NewEncoder(w).Encode([]taskDTO{})
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	start := time.Now()
	end := start.Add(24 * time.Hour)
	_, err := client.FilterTasks(context.Background(), "token", domain.TaskFilter{
		Start: &start,
		End:   &end,
	})
	if err != nil {
		t.Fatalf("FilterTasks() error = %v", err)
	}
}

func TestGetProjectData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("Method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/open/v1/project/p1/data"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		resp := projectDataDTO{
			Project: projectDTO{ID: "p1", Name: "Alpha"},
			Tasks:   []taskDTO{{ID: "t1", Title: "Task 1", ProjectID: "p1"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	project, tasks, err := client.GetProjectData(context.Background(), "token", "p1")
	if err != nil {
		t.Fatalf("GetProjectData() error = %v", err)
	}
	if project.Name != "Alpha" {
		t.Fatalf("Name = %q, want Alpha", project.Name)
	}
	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
}

func TestCreateTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("Method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/open/v1/task"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode error = %v", err)
		}
		if got, want := body["title"], "New Task"; got != want {
			t.Fatalf("title = %v, want %v", got, want)
		}
		resp := taskDTO{ID: "t2", Title: "New Task", ProjectID: "p1", Priority: 0, Status: 0}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	task, err := client.CreateTask(context.Background(), "token", domain.CreateTaskPayload{
		ProjectID: "p1",
		Title:     "New Task",
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if task.Title != "New Task" {
		t.Fatalf("Title = %q, want New Task", task.Title)
	}
}

func TestCreateTaskWithDates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode error = %v", err)
		}
		if body["startDate"] == nil {
			t.Fatal("startDate missing")
		}
		if body["dueDate"] == nil {
			t.Fatal("dueDate missing")
		}
		resp := taskDTO{ID: "t3", Title: "Dated Task", ProjectID: "p1"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	now := time.Now()
	_, err := client.CreateTask(context.Background(), "token", domain.CreateTaskPayload{
		ProjectID: "p1",
		Title:     "Dated Task",
		StartDate: &now,
		DueDate:   &now,
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
}

func TestUpdateTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/open/v1/task/t1"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode error = %v", err)
		}
		if got, want := body["id"], "t1"; got != want {
			t.Fatalf("id = %v, want %v", got, want)
		}
		resp := taskDTO{ID: "t1", Title: "Updated", ProjectID: "p1", Priority: 3, Status: 0}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	task, err := client.UpdateTask(context.Background(), "token", domain.Task{
		ID:        "t1",
		ProjectID: "p1",
		Title:     "Updated",
		Priority:  domain.PriorityMedium,
		Status:    domain.StatusOpen,
	})
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	if task.Title != "Updated" {
		t.Fatalf("Title = %q, want Updated", task.Title)
	}
}

func TestCompleteTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/open/v1/project/p1/task/t1/complete"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	if err := client.CompleteTask(context.Background(), "token", "p1", "t1"); err != nil {
		t.Fatalf("CompleteTask() error = %v", err)
	}
}

func TestDeleteTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodDelete; got != want {
			t.Fatalf("Method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/open/v1/project/p1/task/t1"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	if err := client.DeleteTask(context.Background(), "token", "p1", "t1"); err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}
}

func TestMoveTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("Method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/open/v1/task/move"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		var body []moveRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode error = %v", err)
		}
		if len(body) != 1 {
			t.Fatalf("len(body) = %d, want 1", len(body))
		}
		if got, want := body[0].FromProjectID, "p1"; got != want {
			t.Fatalf("FromProjectID = %q, want %q", got, want)
		}
		if got, want := body[0].ToProjectID, "p2"; got != want {
			t.Fatalf("ToProjectID = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	if err := client.MoveTask(context.Background(), "token", "p1", "p2", "t1"); err != nil {
		t.Fatalf("MoveTask() error = %v", err)
	}
}

func TestFilterTasksReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	_, err := client.FilterTasks(context.Background(), "token", domain.TaskFilter{})
	if err == nil {
		t.Fatal("FilterTasks() error = nil, want error")
	}
}

func TestMapTasks(t *testing.T) {
	dtos := []taskDTO{
		{ID: "t2", Title: "Beta", ProjectID: "p1", Priority: 1, Status: 0},
		{ID: "t1", Title: "Alpha", ProjectID: "p1", Priority: 3, Status: 2},
	}
	tasks := mapTasks(dtos)
	if len(tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(tasks))
	}
	// Verify sorted by title
	if tasks[0].Title != "Alpha" {
		t.Fatalf("Title = %q, want Alpha", tasks[0].Title)
	}
	if tasks[0].Priority != domain.PriorityMedium {
		t.Fatalf("Priority = %v, want medium", tasks[0].Priority)
	}
	if tasks[0].Status != domain.StatusCompleted {
		t.Fatalf("Status = %v, want completed", tasks[0].Status)
	}
	if tasks[1].Status != domain.StatusOpen {
		t.Fatalf("Status = %v, want open", tasks[1].Status)
	}
}

func TestParseTickTimeValid(t *testing.T) {
	result := parseTickTime("2026-05-13T10:00:00.000+0800")
	if result == nil {
		t.Fatal("parseTickTime returned nil, want non-nil")
	}
	if result.Year() != 2026 {
		t.Fatalf("Year = %d, want 2026", result.Year())
	}
}

func TestParseTickTimeAlternativeFormat(t *testing.T) {
	result := parseTickTime("2026-05-13T10:00:00+0800")
	if result == nil {
		t.Fatal("parseTickTime returned nil, want non-nil")
	}
}

func TestParseTickTimeEmpty(t *testing.T) {
	result := parseTickTime("")
	if result != nil {
		t.Fatalf("parseTickTime returned %v, want nil", result)
	}
}

func TestParseTickTimeInvalid(t *testing.T) {
	result := parseTickTime("not-a-date")
	if result != nil {
		t.Fatalf("parseTickTime returned %v, want nil", result)
	}
}
