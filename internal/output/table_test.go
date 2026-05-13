package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/jeely/ticktick-cli/internal/domain"
)

func TestPrintProjectsTable(t *testing.T) {
	var buf bytes.Buffer
	projects := []domain.Project{
		{ID: "p1", Name: "Alpha", Closed: false, Kind: "TASK"},
		{ID: "p2", Name: "Beta", Closed: true, Kind: "NOTE"},
	}
	if err := PrintProjectsTable(&buf, projects); err != nil {
		t.Fatalf("PrintProjectsTable() error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Alpha") {
		t.Fatalf("output missing Alpha: %s", out)
	}
	if !strings.Contains(out, "Beta") {
		t.Fatalf("output missing Beta: %s", out)
	}
}

func TestPrintProjectsTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintProjectsTable(&buf, []domain.Project{}); err != nil {
		t.Fatalf("PrintProjectsTable() error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "ID") {
		t.Fatalf("output missing header: %s", out)
	}
}

func TestPrintTasksTable(t *testing.T) {
	var buf bytes.Buffer
	due := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	tasks := []domain.Task{
		{ID: "t1", Title: "Task 1", ProjectID: "p1", DueDate: &due, Priority: domain.PriorityHigh, Status: domain.StatusOpen},
	}
	names := map[string]string{"p1": "Project A"}
	if err := PrintTasksTable(&buf, tasks, names); err != nil {
		t.Fatalf("PrintTasksTable() error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Task 1") {
		t.Fatalf("output missing Task 1: %s", out)
	}
	if !strings.Contains(out, "Project A") {
		t.Fatalf("output missing project name: %s", out)
	}
}

func TestPrintTasksTableUnknownProject(t *testing.T) {
	var buf bytes.Buffer
	tasks := []domain.Task{
		{ID: "t1", Title: "Task 1", ProjectID: "p-unknown", Priority: domain.PriorityNone, Status: domain.StatusOpen},
	}
	if err := PrintTasksTable(&buf, tasks, map[string]string{}); err != nil {
		t.Fatalf("PrintTasksTable() error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Task 1") {
		t.Fatalf("output missing Task 1: %s", out)
	}
}

func TestPrintFocusTable(t *testing.T) {
	var buf bytes.Buffer
	start := time.Date(2026, 5, 13, 9, 0, 0, 0, time.UTC)
	focuses := []domain.Focus{
		{ID: "f1", Title: "Focus 1", ProjectID: "p1", StartDate: &start, Mode: domain.FocusModeTimer, Status: domain.FocusStatusActive},
	}
	names := map[string]string{"p1": "Project A"}
	if err := PrintFocusTable(&buf, focuses, names); err != nil {
		t.Fatalf("PrintFocusTable() error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Focus 1") {
		t.Fatalf("output missing Focus 1: %s", out)
	}
}

func TestPrintHabitsTable(t *testing.T) {
	var buf bytes.Buffer
	habits := []domain.Habit{
		{ID: "h1", Name: "Read", Goal: 30, CurrentStreak: 5, Status: domain.HabitStatusActive},
	}
	if err := PrintHabitsTable(&buf, habits); err != nil {
		t.Fatalf("PrintHabitsTable() error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Read") {
		t.Fatalf("output missing Read: %s", out)
	}
	if !strings.Contains(out, "active") {
		t.Fatalf("output missing status: %s", out)
	}
}

func TestPrintCheckinsTable(t *testing.T) {
	var buf bytes.Buffer
	checkins := []domain.HabitCheckin{
		{HabitID: "h1", Year: 2026, Stamp: 20260513, Value: 1, Goal: 1},
	}
	if err := PrintCheckinsTable(&buf, checkins); err != nil {
		t.Fatalf("PrintCheckinsTable() error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "2026-05-13") {
		t.Fatalf("output missing formatted date: %s", out)
	}
}

func TestFormatCheckinStampValid(t *testing.T) {
	if got, want := formatCheckinStamp(20260513), "2026-05-13"; got != want {
		t.Fatalf("formatCheckinStamp(20260513) = %q, want %q", got, want)
	}
}

func TestFormatCheckinStampInvalidLength(t *testing.T) {
	if got, want := formatCheckinStamp(123), "123"; got != want {
		t.Fatalf("formatCheckinStamp(123) = %q, want %q", got, want)
	}
}

func TestFormatTimeNil(t *testing.T) {
	if got, want := FormatTime(nil), ""; got != want {
		t.Fatalf("FormatTime(nil) = %q, want %q", got, want)
	}
}

func TestFormatTimeValue(t *testing.T) {
	loc := time.FixedZone("UTC+8", 8*60*60)
	v := time.Date(2026, 5, 13, 10, 30, 0, 0, loc)
	got := FormatTime(&v)
	// Local time depends on test environment, just verify non-empty
	if got == "" {
		t.Fatal("FormatTime(value) = empty, want non-empty")
	}
}
