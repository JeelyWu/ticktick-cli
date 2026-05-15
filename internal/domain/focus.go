package domain

import "time"

type FocusMode int

const (
	FocusModeTimer   FocusMode = 1 // 正计时
	FocusModePomodoro FocusMode = 2 // 番茄钟
)

func (m FocusMode) String() string {
	switch m {
	case FocusModeTimer:
		return "timer"
	case FocusModePomodoro:
		return "pomodoro"
	default:
		return "unknown"
	}
}

func FocusModeFromAPIType(t int) FocusMode {
	switch t {
	case 0:
		return FocusModePomodoro
	case 1:
		return FocusModeTimer
	default:
		return FocusModeTimer
	}
}

func (m FocusMode) APIType() int {
	switch m {
	case FocusModePomodoro:
		return 0
	case FocusModeTimer:
		return 1
	default:
		return 1
	}
}

type FocusStatus int

const (
	FocusStatusActive   FocusStatus = 0 // 进行中
	FocusStatusCompleted FocusStatus = 1 // 已完成
)

func (s FocusStatus) String() string {
	switch s {
	case FocusStatusActive:
		return "active"
	case FocusStatusCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

type Focus struct {
	ID            string      `json:"id"`
	Mode          FocusMode   `json:"mode"`
	Status        FocusStatus `json:"status"`
	Title         string      `json:"title"`
	Content       string      `json:"content"`
	ProjectID     string      `json:"projectId"`
	TaskID        string      `json:"taskId"`
	StartDate     *time.Time  `json:"startDate"`
	EndDate       *time.Time  `json:"endDate"`
	TimeZone      string      `json:"timezone"`
	AbandonReason string      `json:"abandonReason"`
	Tags          []string    `json:"tags"`
	Creators      []string    `json:"creators"`
	SortOrder     int         `json:"sortOrder"`
}

