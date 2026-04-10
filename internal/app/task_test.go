package app

import (
	"testing"
	"time"

	"github.com/jeely/ticktick-cli/internal/domain"
)

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
