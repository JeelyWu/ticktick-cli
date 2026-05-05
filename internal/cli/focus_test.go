package cli

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/jeely/ticktick-cli/internal/domain"
)

type fakeFocusAPI struct {
	focuses  []domain.Focus
	projects []domain.Project
}

func (f fakeFocusAPI) ListProjects(context.Context, string) ([]domain.Project, error) {
	return f.projects, nil
}

func (f fakeFocusAPI) GetFocus(_ context.Context, _ string, focusID string) (domain.Focus, error) {
	for _, f := range f.focuses {
		if f.ID == focusID {
			return f, nil
		}
	}
	return domain.Focus{}, nil
}

func (f fakeFocusAPI) ListFocus(context.Context, string, time.Time, time.Time) ([]domain.Focus, error) {
	return f.focuses, nil
}

func (f fakeFocusAPI) StartFocus(context.Context, string, domain.StartFocusInput) (domain.Focus, error) {
	return domain.Focus{ID: "f1", Title: "Test"}, nil
}

func (f fakeFocusAPI) StopFocus(context.Context, string, string) error {
	return nil
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

func TestFocusStartRequiresProject(t *testing.T) {
	streams, _, _ := newTestStreams()
	cmd := NewFocusCommand(func() (*app.FocusApp, error) {
		return &app.FocusApp{}, nil
	}, nil, streams)
	cmd.SetArgs([]string{"start", "Deep work"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "no project specified") {
		t.Fatalf("error = %q, want no project", err.Error())
	}
}

func TestFocusStopPrintsStopped(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	focusApp := &app.FocusApp{
		Auth:   fakeTokenSource{},
		Client: fakeFocusAPI{},
	}
	cmd := NewRootCommand(RootOptions{
		Version: "dev",
		Streams: streams,
		FocusResolver: func() (*app.FocusApp, error) {
			return focusApp, nil
		},
	})
	cmd.SetArgs([]string{"focus", "stop", "f1"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Stopped") {
		t.Fatalf("stdout = %q, want Stopped", stdout.String())
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
