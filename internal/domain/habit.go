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
	Goal          int         `json:"goal"`
	Color         string      `json:"color"`
	SectionID     string      `json:"sectionId"`
	Icon          string      `json:"icon"`
	Reminders     []string    `json:"reminders"`
	RepeatRule    string      `json:"repeatRule"`
	TargetDays    []int       `json:"targetDays"`
	Unit          string      `json:"unit"`
	Step          int         `json:"step"`
	TotalCheckins int         `json:"totalCheckins"`
	CurrentStreak int         `json:"currentStreak"`
	CreatedTime   *time.Time  `json:"createdTime"`
	ArchivedTime  *time.Time  `json:"archivedTime"`
}

type HabitCheckin struct {
	ID           string     `json:"id"`
	HabitID      string     `json:"habitId"`
	CheckinStamp int        `json:"checkinStamp"`
	CheckinTime  *time.Time `json:"checkinTime"`
	Status       int        `json:"status"`
	Value        int        `json:"value"`
	Goal         int        `json:"goal"`
}

type CreateHabitInput struct {
	Name       string
	Goal       int
	Color      string
	Icon       string
	RepeatRule string
	TargetDays []int
	Unit       string
	Step       int
}

type UpdateHabitInput struct {
	Reference  string
	Name       string
	Goal       int
	Color      string
	Icon       string
	RepeatRule string
	TargetDays []int
	Unit       string
	Step       int
	Status     *HabitStatus
}

type CreateHabitPayload struct {
	Name       string `json:"name"`
	Goal       int    `json:"goal"`
	Color      string `json:"color,omitempty"`
	Icon       string `json:"icon,omitempty"`
	RepeatRule string `json:"repeatRule,omitempty"`
	TargetDays []int  `json:"targetDays,omitempty"`
	Unit       string `json:"unit,omitempty"`
	Step       int    `json:"step,omitempty"`
}
