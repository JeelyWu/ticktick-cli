package domain

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestTaskJSONUsesCamelCase(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	task := Task{
		ID:            "task-id",
		ProjectID:     "project-id",
		Title:         "Test Task",
		Content:       "content",
		Description:   "description",
		StartDate:     &now,
		DueDate:       &now,
		TimeZone:      "Asia/Shanghai",
		IsAllDay:      true,
		Priority:      PriorityHigh,
		Status:        StatusOpen,
		Tags:          []string{"tag1", "tag2"},
		Kind:          "TEXT",
		CompletedTime: &now,
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	jsonStr := string(data)

	wantKeys := []string{
		`"id"`,
		`"projectId"`,
		`"title"`,
		`"content"`,
		`"description"`,
		`"startDate"`,
		`"dueDate"`,
		`"timeZone"`,
		`"isAllDay"`,
		`"priority"`,
		`"status"`,
		`"tags"`,
		`"kind"`,
		`"completedTime"`,
	}
	for _, key := range wantKeys {
		if !strings.Contains(jsonStr, key) {
			t.Errorf("JSON missing key %s; got: %s", key, jsonStr)
		}
	}

	// Ensure no PascalCase keys leak through
	doNotWant := []string{`"ID"`, `"ProjectID"`, `"Title"`, `"IsAllDay"`, `"CompletedTime"`}
	for _, key := range doNotWant {
		if strings.Contains(jsonStr, key) {
			t.Errorf("JSON should not contain PascalCase key %s; got: %s", key, jsonStr)
		}
	}
}

func TestProjectJSONUsesCamelCase(t *testing.T) {
	project := Project{
		ID:         "project-id",
		Name:       "Test Project",
		Color:      "#F18181",
		SortOrder:  42,
		Closed:     true,
		GroupID:    "group-id",
		ViewMode:   "list",
		Permission: "write",
		Kind:       "TASK",
	}

	data, err := json.Marshal(project)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	jsonStr := string(data)

	wantKeys := []string{
		`"id"`,
		`"name"`,
		`"color"`,
		`"sortOrder"`,
		`"closed"`,
		`"groupId"`,
		`"viewMode"`,
		`"permission"`,
		`"kind"`,
	}
	for _, key := range wantKeys {
		if !strings.Contains(jsonStr, key) {
			t.Errorf("JSON missing key %s; got: %s", key, jsonStr)
		}
	}

	doNotWant := []string{`"ID"`, `"Name"`, `"SortOrder"`, `"GroupID"`}
	for _, key := range doNotWant {
		if strings.Contains(jsonStr, key) {
			t.Errorf("JSON should not contain PascalCase key %s; got: %s", key, jsonStr)
		}
	}
}
