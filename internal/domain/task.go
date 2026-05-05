package domain

import "time"

type Priority int

const (
	PriorityNone   Priority = 0
	PriorityLow    Priority = 1
	PriorityMedium Priority = 3
	PriorityHigh   Priority = 5
)

func (p Priority) String() string {
	switch p {
	case PriorityNone:
		return "none"
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	default:
		return "unknown"
	}
}

type TaskStatus int

const (
	StatusOpen      TaskStatus = 0
	StatusCompleted TaskStatus = 2
)

func (s TaskStatus) String() string {
	switch s {
	case StatusOpen:
		return "open"
	case StatusCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

type Task struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"projectId"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	Description   string     `json:"description"`
	StartDate     *time.Time `json:"startDate"`
	DueDate       *time.Time `json:"dueDate"`
	TimeZone      string     `json:"timeZone"`
	IsAllDay      bool       `json:"isAllDay"`
	Priority      Priority   `json:"priority"`
	Status        TaskStatus `json:"status"`
	Tags          []string   `json:"tags"`
	Kind          string     `json:"kind"`
	CompletedTime *time.Time `json:"completedTime"`
}

type CreateTaskInput struct {
	ProjectRef  string
	Title       string
	Content     string
	Description string
	StartRaw    string
	DueRaw      string
	AllDay      bool
	Priority    Priority
}

type UpdateTaskInput struct {
	Reference   string
	ProjectRef  string
	Title       string
	Content     string
	Description string
	StartRaw    string
	DueRaw      string
	AllDay      *bool
	Priority    *Priority
}

type MoveTaskInput struct {
	Reference      string
	FromProjectRef string
	ToProjectRef   string
}

type CreateTaskPayload struct {
	ProjectID   string
	Title       string
	Content     string
	Description string
	StartDate   *time.Time
	DueDate     *time.Time
	AllDay      bool
	Priority    Priority
}
