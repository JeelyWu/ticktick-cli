package app

import (
	"context"
	"testing"
	"time"

	"github.com/jeelywu/ticktick-cli/internal/domain"
)

type stubTokenSource struct{}

func (stubTokenSource) AccessToken(context.Context) (string, error) {
	return "token-1", nil
}

type recordingTaskAPI struct {
	projects    []domain.Project
	filterTasks []domain.Task
	lastFilter  domain.TaskFilter

	updateCalls []domain.Task
	deleteCalls []deleteCall
	moveCalls   []moveCall
}

type deleteCall struct {
	ProjectID string
	TaskID    string
}

type moveCall struct {
	FromProjectID string
	ToProjectID   string
	TaskID        string
}

func (r *recordingTaskAPI) ListProjects(context.Context, string) ([]domain.Project, error) {
	if r.projects != nil {
		return r.projects, nil
	}
	return []domain.Project{{ID: "p1", Name: "Zipto"}}, nil
}

func (r *recordingTaskAPI) FilterTasks(_ context.Context, _ string, filter domain.TaskFilter) ([]domain.Task, error) {
	r.lastFilter = filter
	if len(filter.Statuses) == 0 {
		return r.filterTasks, nil
	}
	allowed := make(map[domain.TaskStatus]struct{}, len(filter.Statuses))
	for _, s := range filter.Statuses {
		allowed[s] = struct{}{}
	}
	out := make([]domain.Task, 0, len(r.filterTasks))
	for _, task := range r.filterTasks {
		if _, ok := allowed[task.Status]; ok {
			out = append(out, task)
		}
	}
	return out, nil
}

func (r *recordingTaskAPI) GetProjectData(context.Context, string, string) (domain.Project, []domain.Task, error) {
	return domain.Project{}, nil, nil
}

func (r *recordingTaskAPI) CreateTask(context.Context, string, domain.CreateTaskPayload) (domain.Task, error) {
	return domain.Task{}, nil
}

func (r *recordingTaskAPI) UpdateTask(_ context.Context, _ string, task domain.Task) (domain.Task, error) {
	r.updateCalls = append(r.updateCalls, task)
	return task, nil
}

func (r *recordingTaskAPI) CompleteTask(context.Context, string, string, string) error {
	return nil
}

func (r *recordingTaskAPI) DeleteTask(_ context.Context, _ string, projectID, taskID string) error {
	r.deleteCalls = append(r.deleteCalls, deleteCall{ProjectID: projectID, TaskID: taskID})
	return nil
}

func (r *recordingTaskAPI) MoveTask(_ context.Context, _ string, fromProjectID, toProjectID, taskID string) error {
	r.moveCalls = append(r.moveCalls, moveCall{FromProjectID: fromProjectID, ToProjectID: toProjectID, TaskID: taskID})
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

func TestTaskAppTodayRequestsOpenTasks(t *testing.T) {
	now := time.Date(2026, 4, 9, 9, 0, 0, 0, time.Local)
	due := time.Date(2026, 4, 9, 17, 0, 0, 0, time.Local)
	client := &recordingTaskAPI{
		filterTasks: []domain.Task{
			{ID: "today", Title: "Today", ProjectID: "p1", Status: domain.StatusOpen, DueDate: &due},
			{ID: "completed", Title: "Done", ProjectID: "p1", Status: domain.StatusCompleted, DueDate: &due},
		},
	}
	taskApp := TaskApp{
		Auth:   stubTokenSource{},
		Client: client,
		Now: func() time.Time {
			return now
		},
	}

	tasks, _, err := taskApp.Today(context.Background())
	if err != nil {
		t.Fatalf("Today() error = %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != "today" {
		t.Fatalf("tasks = %#v, want only open today task", tasks)
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

func TestTaskAppRemoveFindsCompleted(t *testing.T) {
	client := &recordingTaskAPI{
		filterTasks: []domain.Task{
			{ID: "t1", Title: "Foo", ProjectID: "p1", Status: domain.StatusCompleted},
		},
	}
	taskApp := TaskApp{Auth: stubTokenSource{}, Client: client}

	if err := taskApp.Remove(context.Background(), "Foo", ""); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if len(client.deleteCalls) != 1 {
		t.Fatalf("deleteCalls = %d, want 1", len(client.deleteCalls))
	}
	if client.deleteCalls[0].TaskID != "t1" {
		t.Fatalf("deleted task = %s, want t1", client.deleteCalls[0].TaskID)
	}
}

func TestTaskAppUpdateFindsCompleted(t *testing.T) {
	client := &recordingTaskAPI{
		filterTasks: []domain.Task{
			{ID: "t1", Title: "Foo", ProjectID: "p1", Status: domain.StatusCompleted},
		},
	}
	taskApp := TaskApp{Auth: stubTokenSource{}, Client: client}

	_, err := taskApp.Update(context.Background(), domain.UpdateTaskInput{
		Reference: "Foo",
		Title:     "Bar",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if len(client.updateCalls) != 1 {
		t.Fatalf("updateCalls = %d, want 1", len(client.updateCalls))
	}
	if client.updateCalls[0].Title != "Bar" {
		t.Fatalf("updated title = %s, want Bar", client.updateCalls[0].Title)
	}
}

func TestTaskAppMoveFindsCompleted(t *testing.T) {
	client := &recordingTaskAPI{
		projects: []domain.Project{
			{ID: "p1", Name: "Src"},
			{ID: "p2", Name: "Dst"},
		},
		filterTasks: []domain.Task{
			{ID: "t1", Title: "Foo", ProjectID: "p1", Status: domain.StatusCompleted},
		},
	}
	taskApp := TaskApp{Auth: stubTokenSource{}, Client: client}

	if err := taskApp.Move(context.Background(), domain.MoveTaskInput{
		Reference:      "Foo",
		FromProjectRef: "Src",
		ToProjectRef:   "Dst",
	}); err != nil {
		t.Fatalf("Move() error = %v", err)
	}
	if len(client.moveCalls) != 1 {
		t.Fatalf("moveCalls = %d, want 1", len(client.moveCalls))
	}
	if client.moveCalls[0].TaskID != "t1" {
		t.Fatalf("moved task = %s, want t1", client.moveCalls[0].TaskID)
	}
}

func TestTaskAppReopenSetsStatusOpen(t *testing.T) {
	client := &recordingTaskAPI{
		filterTasks: []domain.Task{
			{ID: "t1", Title: "Foo", ProjectID: "p1", Status: domain.StatusCompleted},
		},
	}
	taskApp := TaskApp{Auth: stubTokenSource{}, Client: client}

	if err := taskApp.Reopen(context.Background(), "Foo", ""); err != nil {
		t.Fatalf("Reopen() error = %v", err)
	}
	if len(client.updateCalls) != 1 {
		t.Fatalf("updateCalls = %d, want 1", len(client.updateCalls))
	}
	if client.updateCalls[0].Status != domain.StatusOpen {
		t.Fatalf("status = %v, want open", client.updateCalls[0].Status)
	}
	if client.updateCalls[0].ID != "t1" {
		t.Fatalf("updated task = %s, want t1", client.updateCalls[0].ID)
	}
}
