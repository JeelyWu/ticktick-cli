package cli

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/jeely/ticktick-cli/internal/config"
	"github.com/jeely/ticktick-cli/internal/domain"
)

type fakeTokenSource struct{}

func (fakeTokenSource) AccessToken(context.Context) (string, error) {
	return "token", nil
}

type fakeTaskAPI struct {
	projects []domain.Project
	tasks    []domain.Task
}

func (f fakeTaskAPI) ListProjects(context.Context, string) ([]domain.Project, error) {
	return f.projects, nil
}

func (f fakeTaskAPI) FilterTasks(context.Context, string, domain.TaskFilter) ([]domain.Task, error) {
	return f.tasks, nil
}

func (f fakeTaskAPI) GetProjectData(context.Context, string, string) (domain.Project, []domain.Task, error) {
	return domain.Project{}, f.tasks, nil
}

func (f fakeTaskAPI) CreateTask(context.Context, string, domain.CreateTaskPayload) (domain.Task, error) {
	return domain.Task{}, nil
}

func (f fakeTaskAPI) UpdateTask(context.Context, string, domain.Task) (domain.Task, error) {
	return domain.Task{}, nil
}

func (f fakeTaskAPI) CompleteTask(context.Context, string, string, string) error {
	return nil
}

func (f fakeTaskAPI) DeleteTask(context.Context, string, string, string) error {
	return nil
}

func (f fakeTaskAPI) MoveTask(context.Context, string, string, string, string) error {
	return nil
}

type fakeProjectAPI struct {
	projects []domain.Project
}

func (f fakeProjectAPI) ListProjects(context.Context, string) ([]domain.Project, error) {
	return f.projects, nil
}

func (f fakeProjectAPI) CreateProject(context.Context, string, string, string, string) (domain.Project, error) {
	return domain.Project{}, nil
}

func (f fakeProjectAPI) UpdateProject(context.Context, string, string, string, string, string) (domain.Project, error) {
	return domain.Project{}, nil
}

func (f fakeProjectAPI) DeleteProject(context.Context, string, string) error {
	return nil
}

func testConfigResolver(t *testing.T, output string) ConfigResolver {
	t.Helper()
	store := config.NewStore(t.TempDir() + "/config.yaml")
	configApp := &app.ConfigApp{Store: store}
	if err := configApp.Set(context.Background(), "output.default", output); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	return func() (*app.ConfigApp, error) {
		return &app.ConfigApp{Store: store}, nil
	}
}

func TestTodayCommandUsesConfiguredJSONOutput(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	project := domain.Project{ID: "p1", Name: "Inbox"}
	due := time.Date(2026, 4, 17, 9, 0, 0, 0, time.Local)
	taskApp := &app.TaskApp{
		Auth:   fakeTokenSource{},
		Client: fakeTaskAPI{projects: []domain.Project{project}, tasks: []domain.Task{{ID: "t1", ProjectID: "p1", Title: "Ship it", DueDate: &due, Priority: domain.PriorityHigh, Status: domain.StatusOpen}}},
		Now: func() time.Time {
			return due
		},
	}
	cmd := NewRootCommand(RootOptions{
		Version: "dev",
		Streams: streams,
		TaskResolver: func() (*app.TaskApp, error) {
			return taskApp, nil
		},
		ConfigResolver: testConfigResolver(t, "json"),
	})
	cmd.SetArgs([]string{"today"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "\"ID\": \"t1\"") {
		t.Fatalf("stdout = %q, want JSON output", stdout.String())
	}
	if strings.HasPrefix(stdout.String(), "ID") {
		t.Fatalf("stdout = %q, want configured JSON instead of table", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestInboxCommandUsesConfiguredJSONOutput(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	project := domain.Project{ID: "p1", Name: "Inbox"}
	taskApp := &app.TaskApp{
		Auth:   fakeTokenSource{},
		Client: fakeTaskAPI{projects: []domain.Project{project}, tasks: []domain.Task{{ID: "t1", ProjectID: "p1", Title: "Ship it", Priority: domain.PriorityHigh, Status: domain.StatusOpen}}},
	}
	cmd := NewRootCommand(RootOptions{
		Version: "dev",
		Streams: streams,
		TaskResolver: func() (*app.TaskApp, error) {
			return taskApp, nil
		},
		ConfigResolver: testConfigResolver(t, "json"),
	})
	cmd.SetArgs([]string{"inbox"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "\"ID\": \"t1\"") {
		t.Fatalf("stdout = %q, want JSON output", stdout.String())
	}
	if strings.HasPrefix(stdout.String(), "ID") {
		t.Fatalf("stdout = %q, want configured JSON instead of table", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestTaskListCommandFlagOverridesConfiguredJSONOutput(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	project := domain.Project{ID: "p1", Name: "Inbox"}
	taskApp := &app.TaskApp{
		Auth:   fakeTokenSource{},
		Client: fakeTaskAPI{projects: []domain.Project{project}, tasks: []domain.Task{{ID: "t1", ProjectID: "p1", Title: "Ship it", Priority: domain.PriorityHigh, Status: domain.StatusOpen}}},
	}
	cmd := NewRootCommand(RootOptions{
		Version: "dev",
		Streams: streams,
		TaskResolver: func() (*app.TaskApp, error) {
			return taskApp, nil
		},
		ConfigResolver: testConfigResolver(t, "json"),
	})
	cmd.SetArgs([]string{"task", "ls", "--output", "table"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.HasPrefix(stdout.String(), "ID") {
		t.Fatalf("stdout = %q, want table output", stdout.String())
	}
	if strings.Contains(stdout.String(), "\"ID\": \"t1\"") {
		t.Fatalf("stdout = %q, want flag to override configured JSON", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestTaskGetCommandUsesConfiguredJSONOutput(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	project := domain.Project{ID: "p1", Name: "Inbox"}
	taskApp := &app.TaskApp{
		Auth:   fakeTokenSource{},
		Client: fakeTaskAPI{projects: []domain.Project{project}, tasks: []domain.Task{{ID: "t1", ProjectID: "p1", Title: "Ship it", Priority: domain.PriorityHigh, Status: domain.StatusOpen}}},
	}
	cmd := NewRootCommand(RootOptions{
		Version: "dev",
		Streams: streams,
		TaskResolver: func() (*app.TaskApp, error) {
			return taskApp, nil
		},
		ConfigResolver: testConfigResolver(t, "json"),
	})
	cmd.SetArgs([]string{"task", "get", "t1"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "\"ID\": \"t1\"") {
		t.Fatalf("stdout = %q, want JSON output", stdout.String())
	}
	if strings.HasPrefix(stdout.String(), "ID") {
		t.Fatalf("stdout = %q, want configured JSON instead of table", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestProjectListCommandUsesConfiguredJSONOutput(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	cmd := NewRootCommand(RootOptions{
		Version: "dev",
		Streams: streams,
		ProjectResolver: func() (*app.ProjectApp, error) {
			return &app.ProjectApp{
				Auth:   fakeTokenSource{},
				Client: fakeProjectAPI{projects: []domain.Project{{ID: "p1", Name: "Inbox", Kind: "TASK"}}},
			}, nil
		},
		ConfigResolver: testConfigResolver(t, "json"),
	})
	cmd.SetArgs([]string{"project", "ls"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "\"ID\": \"p1\"") {
		t.Fatalf("stdout = %q, want JSON output", stdout.String())
	}
	if strings.HasPrefix(stdout.String(), "ID") {
		t.Fatalf("stdout = %q, want configured JSON instead of table", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestProjectGetCommandUsesConfiguredJSONOutput(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	cmd := NewRootCommand(RootOptions{
		Version: "dev",
		Streams: streams,
		ProjectResolver: func() (*app.ProjectApp, error) {
			return &app.ProjectApp{
				Auth:   fakeTokenSource{},
				Client: fakeProjectAPI{projects: []domain.Project{{ID: "p1", Name: "Inbox", Kind: "TASK"}}},
			}, nil
		},
		ConfigResolver: testConfigResolver(t, "json"),
	})
	cmd.SetArgs([]string{"project", "get", "p1"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "\"ID\": \"p1\"") {
		t.Fatalf("stdout = %q, want JSON output", stdout.String())
	}
	if strings.HasPrefix(stdout.String(), "ID") {
		t.Fatalf("stdout = %q, want configured JSON instead of table", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}
