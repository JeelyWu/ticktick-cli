package domain

import "time"

type TaskFilter struct {
	ProjectIDs []string
	Tags       []string
	Statuses   []TaskStatus
	Priorities []Priority
	Start      *time.Time
	End        *time.Time
}

func (f TaskFilter) StatusCodes() []int {
	out := make([]int, 0, len(f.Statuses))
	for _, status := range f.Statuses {
		out = append(out, int(status))
	}
	return out
}

func (f TaskFilter) PriorityCodes() []int {
	out := make([]int, 0, len(f.Priorities))
	for _, priority := range f.Priorities {
		out = append(out, int(priority))
	}
	return out
}
