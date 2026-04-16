package cli

import (
	"strings"
	"testing"

	"github.com/jeely/ticktick-cli/internal/app"
)

func TestTaskListRejectsTodayWithCompletedStatus(t *testing.T) {
	streams, _, _ := newTestStreams()
	resolved := 0
	cmd := NewTaskCommand(func() (*app.TaskApp, error) {
		resolved++
		return &app.TaskApp{}, nil
	}, nil, streams)
	cmd.SetArgs([]string{"ls", "--today", "--status", "completed"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "--today requires open tasks") {
		t.Fatalf("error = %q, want today/open conflict", err)
	}
	if resolved != 0 {
		t.Fatalf("resolver calls = %d, want 0", resolved)
	}
}

func TestTaskListRejectsOverdueWithCompletedStatus(t *testing.T) {
	streams, _, _ := newTestStreams()
	resolved := 0
	cmd := NewTaskCommand(func() (*app.TaskApp, error) {
		resolved++
		return &app.TaskApp{}, nil
	}, nil, streams)
	cmd.SetArgs([]string{"ls", "--overdue", "--status", "completed"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "--overdue requires open tasks") {
		t.Fatalf("error = %q, want overdue/open conflict", err)
	}
	if resolved != 0 {
		t.Fatalf("resolver calls = %d, want 0", resolved)
	}
}

func TestTaskListRejectsTodayAndOverdueTogether(t *testing.T) {
	streams, _, _ := newTestStreams()
	resolved := 0
	cmd := NewTaskCommand(func() (*app.TaskApp, error) {
		resolved++
		return &app.TaskApp{}, nil
	}, nil, streams)
	cmd.SetArgs([]string{"ls", "--today", "--overdue"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "--today and --overdue cannot be used together") {
		t.Fatalf("error = %q, want mutual exclusion", err)
	}
	if resolved != 0 {
		t.Fatalf("resolver calls = %d, want 0", resolved)
	}
}
