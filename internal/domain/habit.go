package domain

import "time"

type HabitStatus int

const (
	HabitStatusActive   HabitStatus = 0
	HabitStatusArchived HabitStatus = 1
)

func (s HabitStatus) String() string {
	switch s {
	case HabitStatusActive:
		return "active"
	case HabitStatusArchived:
		return "archived"
	default:
		return "unknown"
	}
}

type Habit struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Status        HabitStatus `json:"status"`
	Goal          float64     `json:"goal"`
	Color         string      `json:"color"`
	SectionID     string      `json:"sectionId"`
	IconRes       string      `json:"iconRes"`
	Reminders     []string    `json:"reminders"`
	RepeatRule    string      `json:"repeatRule"`
	TargetDays    int         `json:"targetDays"`
	Unit          string      `json:"unit"`
	Step          float64     `json:"step"`
	TotalCheckins int         `json:"totalCheckins"`
	CurrentStreak int         `json:"currentStreak"`
	CreatedTime   *time.Time  `json:"createdTime"`
	ArchivedTime  *time.Time  `json:"archivedTime"`
}

type HabitCheckin struct {
	HabitID string  `json:"habitId"`
	Year    int     `json:"year"`
	Stamp   int     `json:"stamp"`
	Value   float64 `json:"value"`
	Goal    float64 `json:"goal"`
}

type CreateHabitInput struct {
	Name       string
	Goal       float64
	Color      string
	IconRes    string
	RepeatRule string
	TargetDays int
	Unit       string
	Step       float64
}

type UpdateHabitInput struct {
	Reference  string
	Name       string
	Goal       float64
	Color      string
	IconRes    string
	RepeatRule string
	TargetDays int
	Unit       string
	Step       float64
	Status     *HabitStatus
}

type CreateHabitPayload struct {
	Name       string    `json:"name"`
	Goal       float64   `json:"goal"`
	Color      string    `json:"color,omitempty"`
	IconRes    string    `json:"iconRes,omitempty"`
	RepeatRule string    `json:"repeatRule,omitempty"`
	TargetDays int       `json:"targetDays,omitempty"`
	Unit       string    `json:"unit,omitempty"`
	Step       float64   `json:"step,omitempty"`
	Status     int       `json:"status,omitempty"`
	StatusSet  bool
}
