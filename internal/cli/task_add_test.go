package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/jeelywu/ticktick-cli/internal/app"
	"github.com/jeelywu/ticktick-cli/internal/config"
	"github.com/jeelywu/ticktick-cli/internal/domain"
)

func TestTaskAddNonTerminalWithoutProjectFails(t *testing.T) {
	streams, _, _ := newTestStreams()
	cmd := NewTaskCommand(func() (*app.TaskApp, error) {
		return &app.TaskApp{}, nil
	}, nil, streams)
	cmd.SetArgs([]string{"add", "foo"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "no project specified") {
		t.Fatalf("error = %q, want no project specified", err.Error())
	}
}

type stubTaskAPIWithProjects struct{}

func (stubTaskAPIWithProjects) ListProjects(context.Context, string) ([]domain.Project, error) {
	return []domain.Project{{ID: "p1", Name: "Work", SortOrder: 1}}, nil
}
func (stubTaskAPIWithProjects) FilterTasks(context.Context, string, domain.TaskFilter) ([]domain.Task, error) {
	return nil, nil
}
func (stubTaskAPIWithProjects) GetProjectData(context.Context, string, string) (domain.Project, []domain.Task, error) {
	return domain.Project{}, nil, nil
}
func (stubTaskAPIWithProjects) CreateTask(context.Context, string, domain.CreateTaskPayload) (domain.Task, error) {
	return domain.Task{ID: "t1", Title: "foo", ProjectID: "p1"}, nil
}
func (stubTaskAPIWithProjects) UpdateTask(context.Context, string, domain.Task) (domain.Task, error) {
	return domain.Task{}, nil
}
func (stubTaskAPIWithProjects) CompleteTask(context.Context, string, string, string) error {
	return nil
}
func (stubTaskAPIWithProjects) DeleteTask(context.Context, string, string, string) error {
	return nil
}
func (stubTaskAPIWithProjects) MoveTask(context.Context, string, string, string, string) error {
	return nil
}

type stubTokenSource struct{}

func (stubTokenSource) AccessToken(context.Context) (string, error) {
	return "token", nil
}

func TestTaskAddUsesDefaultProject(t *testing.T) {
	streams, stdout, _ := newTestStreams()
	tmpDir := t.TempDir()
	store := config.NewStore(tmpDir + "/config.yaml")
	var cfg config.Config
	cfg.DefaultProject = "Work"
	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	cmd := NewTaskCommand(
		func() (*app.TaskApp, error) {
			return &app.TaskApp{
				Auth:   stubTokenSource{},
				Client: stubTaskAPIWithProjects{},
			}, nil
		},
		func() (*app.ConfigApp, error) {
			return &app.ConfigApp{Store: store}, nil
		},
		streams,
	)
	cmd.SetArgs([]string{"add", "foo"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "t1") {
		t.Fatalf("output = %q, want task ID t1", stdout.String())
	}
}

func TestSelectProject(t *testing.T) {
	projects := []domain.Project{
		{ID: "p1", Name: "Work"},
		{ID: "p2", Name: "Personal"},
	}
	streams := Streams{
		In:  bytes.NewBufferString("2\n"),
		Out: &bytes.Buffer{},
	}

	project, err := SelectProject(streams, projects)
	if err != nil {
		t.Fatalf("SelectProject() error = %v", err)
	}
	if project.Name != "Personal" {
		t.Fatalf("project.Name = %q, want Personal", project.Name)
	}
}

func TestSelectProjectInvalidThenValid(t *testing.T) {
	projects := []domain.Project{
		{ID: "p1", Name: "Work"},
	}
	streams := Streams{
		In:  bytes.NewBufferString("abc\n0\n2\n1\n"),
		Out: &bytes.Buffer{},
	}

	project, err := SelectProject(streams, projects)
	if err != nil {
		t.Fatalf("SelectProject() error = %v", err)
	}
	if project.Name != "Work" {
		t.Fatalf("project.Name = %q, want Work", project.Name)
	}
}
