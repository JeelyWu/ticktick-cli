package app

import (
	"context"
	"testing"
	"time"

	"github.com/jeely/ticktick-cli/internal/domain"
)

type stubTokenSource struct{}

func (stubTokenSource) AccessToken(context.Context) (string, error) {
	return "token-1", nil
}

type recordingTaskAPI struct {
	filterTasks []domain.Task
	lastFilter  domain.TaskFilter
}

func (r *recordingTaskAPI) ListProjects(context.Context, string) ([]domain.Project, error) {
	return []domain.Project{{ID: "p1", Name: "Zipto"}}, nil
}

func (r *recordingTaskAPI) FilterTasks(_ context.Context, _ string, filter domain.TaskFilter) ([]domain.Task, error) {
	r.lastFilter = filter
	return r.filterTasks, nil
}

func (r *recordingTaskAPI) GetProjectData(context.Context, string, string) (domain.Project, []domain.Task, error) {
	return domain.Project{}, nil, nil
}

func (r *recordingTaskAPI) CreateTask(context.Context, string, domain.CreateTaskPayload) (domain.Task, error) {
	return domain.Task{}, nil
}

func (r *recordingTaskAPI) UpdateTask(context.Context, string, domain.Task) (domain.Task, error) {
	return domain.Task{}, nil
}

func (r *recordingTaskAPI) CompleteTask(context.Context, string, string, string) error {
	return nil
}

func (r *recordingTaskAPI) DeleteTask(context.Context, string, string, string) error {
	return nil
}

func (r *recordingTaskAPI) MoveTask(context.Context, string, string, string, string) error {
	return nil
}

func TestResolveTaskReferenceByID(t *testing.T) {
	tasks := []domain.Task{
		{ID: "a1", Title: "Spec"},
		{ID: "b2", Title: "Review"},
	}

	task, err := resolveTaskReference("b2", tasks)
	if err != nil {
		t.Fatalf("resolveTaskReference() error = %v", err)
	}
	if task.Title != "Review" {
		t.Fatalf("task.Title = %q, want Review", task.Title)
	}
}

func TestTaskIsDueTodayOrOverdue(t *testing.T) {
	now := time.Date(2026, 4, 9, 9, 0, 0, 0, time.Local)
	due := time.Date(2026, 4, 8, 18, 0, 0, 0, time.Local)
	task := domain.Task{DueDate: &due, Status: domain.StatusOpen}

	if !taskIsDueTodayOrOverdue(task, now) {
		t.Fatal("taskIsDueTodayOrOverdue() = false, want true")
	}
}

func TestTaskIsOverdue(t *testing.T) {
	now := time.Date(2026, 4, 9, 9, 0, 0, 0, time.Local)
	due := time.Date(2026, 4, 8, 18, 0, 0, 0, time.Local)
	task := domain.Task{DueDate: &due, Status: domain.StatusOpen}

	if !taskIsOverdue(task, now) {
		t.Fatal("taskIsOverdue() = false, want true")
	}
}

func TestTaskAppListFiltersToday(t *testing.T) {
	now := time.Date(2026, 4, 9, 9, 0, 0, 0, time.Local)
	yesterday := time.Date(2026, 4, 8, 18, 0, 0, 0, time.Local)
	today := time.Date(2026, 4, 9, 17, 0, 0, 0, time.Local)
	tomorrow := time.Date(2026, 4, 10, 9, 0, 0, 0, time.Local)
	client := &recordingTaskAPI{
		filterTasks: []domain.Task{
			{ID: "overdue", Title: "Overdue", ProjectID: "p1", Status: domain.StatusOpen, DueDate: &yesterday},
			{ID: "today", Title: "Today", ProjectID: "p1", Status: domain.StatusOpen, DueDate: &today},
			{ID: "future", Title: "Future", ProjectID: "p1", Status: domain.StatusOpen, DueDate: &tomorrow},
			{ID: "completed", Title: "Done", ProjectID: "p1", Status: domain.StatusCompleted, DueDate: &yesterday},
		},
	}
	taskApp := TaskApp{
		Auth:   stubTokenSource{},
		Client: client,
		Now: func() time.Time {
			return now
		},
	}

	tasks, _, err := taskApp.List(context.Background(), ListTasksInput{
		Statuses: []domain.TaskStatus{domain.StatusOpen},
		Today:    true,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(tasks))
	}
	if tasks[0].ID != "overdue" || tasks[1].ID != "today" {
		t.Fatalf("tasks = %#v, want overdue then today", tasks)
	}
	if got := client.lastFilter.StatusCodes(); len(got) != 1 || got[0] != int(domain.StatusOpen) {
		t.Fatalf("lastFilter.StatusCodes() = %v, want [%d]", got, domain.StatusOpen)
	}
}

func TestTaskAppListFiltersOverdue(t *testing.T) {
	now := time.Date(2026, 4, 9, 9, 0, 0, 0, time.Local)
	yesterday := time.Date(2026, 4, 8, 18, 0, 0, 0, time.Local)
	today := time.Date(2026, 4, 9, 17, 0, 0, 0, time.Local)
	client := &recordingTaskAPI{
		filterTasks: []domain.Task{
			{ID: "overdue", Title: "Overdue", ProjectID: "p1", Status: domain.StatusOpen, DueDate: &yesterday},
			{ID: "today", Title: "Today", ProjectID: "p1", Status: domain.StatusOpen, DueDate: &today},
		},
	}
	taskApp := TaskApp{
		Auth:   stubTokenSource{},
		Client: client,
		Now: func() time.Time {
			return now
		},
	}

	tasks, _, err := taskApp.List(context.Background(), ListTasksInput{
		Statuses: []domain.TaskStatus{domain.StatusOpen},
		Overdue:  true,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	if tasks[0].ID != "overdue" {
		t.Fatalf("tasks[0].ID = %q, want overdue", tasks[0].ID)
	}
}
