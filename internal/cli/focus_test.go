package cli

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jeelywu/ticktick-cli/internal/app"
	"github.com/jeelywu/ticktick-cli/internal/domain"
)

type fakeFocusAPI struct {
	focuses  []domain.Focus
	projects []domain.Project
}

func (f fakeFocusAPI) ListProjects(context.Context, string) ([]domain.Project, error) {
	return f.projects, nil
}

func (f fakeFocusAPI) GetFocus(_ context.Context, _ string, focusID string, focusType int) (domain.Focus, error) {
	for _, f := range f.focuses {
		if f.ID == focusID {
			return f, nil
		}
	}
	return domain.Focus{}, nil
}

func (f fakeFocusAPI) ListFocus(context.Context, string, time.Time, time.Time, int) ([]domain.Focus, error) {
	return f.focuses, nil
}

func TestFocusListPrintsTable(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	project := domain.Project{ID: "p1", Name: "Inbox"}
	focusApp := &app.FocusApp{
		Auth: fakeTokenSource{},
		Client: fakeFocusAPI{
			projects: []domain.Project{project},
			focuses:  []domain.Focus{{ID: "f1", Title: "Deep work", ProjectID: "p1", Mode: domain.FocusModeTimer, Status: domain.FocusStatusActive}},
		},
	}
	cmd := NewRootCommand(RootOptions{
		Version: "dev",
		Streams: streams,
		FocusResolver: func() (*app.FocusApp, error) {
			return focusApp, nil
		},
	})
	cmd.SetArgs([]string{"focus", "ls"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Deep work") {
		t.Fatalf("stdout = %q, want Deep work", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestFocusListJSONOutput(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	project := domain.Project{ID: "p1", Name: "Inbox"}
	focusApp := &app.FocusApp{
		Auth: fakeTokenSource{},
		Client: fakeFocusAPI{
			projects: []domain.Project{project},
			focuses:  []domain.Focus{{ID: "f1", Title: "Deep work", ProjectID: "p1"}},
		},
	}
	cmd := NewRootCommand(RootOptions{
		Version: "dev",
		Streams: streams,
		FocusResolver: func() (*app.FocusApp, error) {
			return focusApp, nil
		},
	})
	cmd.SetArgs([]string{"focus", "ls", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "\"id\": \"f1\"") {
		t.Fatalf("stdout = %q, want JSON output", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestFocusGetPrintsTable(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	focusApp := &app.FocusApp{
		Auth: fakeTokenSource{},
		Client: fakeFocusAPI{
			focuses: []domain.Focus{{ID: "f1", Title: "Deep work", ProjectID: "p1"}},
		},
	}
	cmd := NewRootCommand(RootOptions{
		Version: "dev",
		Streams: streams,
		FocusResolver: func() (*app.FocusApp, error) {
			return focusApp, nil
		},
	})
	cmd.SetArgs([]string{"focus", "get", "f1"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Deep work") {
		t.Fatalf("stdout = %q, want Deep work", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestFocusCommandHelpDoesNotResolve(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	resolved := 0
	cmd := NewRootCommand(RootOptions{
		Version: "dev",
		Streams: streams,
		FocusResolver: func() (*app.FocusApp, error) {
			resolved++
			return nil, nil
		},
	})
	cmd.SetArgs([]string{"focus", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resolved != 0 {
		t.Fatalf("resolver calls = %d, want 0", resolved)
	}
	if !strings.Contains(stdout.String(), "focus") {
		t.Fatalf("help output = %q, want focus command", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}
