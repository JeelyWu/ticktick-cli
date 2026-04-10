package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jeely/ticktick-cli/internal/config"
	"github.com/jeely/ticktick-cli/internal/domain"
)

func ParseQuickAdd(raw string) (domain.CreateTaskInput, error) {
	fields := strings.Fields(raw)
	var out domain.CreateTaskInput
	var title []string

	for _, field := range fields {
		switch {
		case strings.HasPrefix(field, "#"):
			out.ProjectRef = strings.TrimPrefix(field, "#")
		case strings.HasPrefix(field, "!"):
			value, err := strconv.Atoi(strings.TrimPrefix(field, "!"))
			if err != nil {
				return domain.CreateTaskInput{}, err
			}
			out.Priority = domain.Priority(value)
		case strings.HasPrefix(field, "^"):
			out.DueRaw = strings.TrimPrefix(field, "^")
		default:
			title = append(title, field)
		}
	}

	if len(title) == 0 {
		return domain.CreateTaskInput{}, fmt.Errorf("quick add title is required")
	}
	out.Title = strings.Join(title, " ")
	return out, nil
}

type QuickAddApp struct {
	TaskApp     *TaskApp
	ConfigStore *config.Store
}

func (a QuickAddApp) Add(ctx context.Context, raw string) (domain.Task, error) {
	input, err := ParseQuickAdd(raw)
	if err != nil {
		return domain.Task{}, err
	}
	if input.ProjectRef == "" && a.ConfigStore != nil {
		cfg, err := a.ConfigStore.Load()
		if err != nil {
			return domain.Task{}, err
		}
		input.ProjectRef = cfg.Task.DefaultProject
	}
	return a.TaskApp.Add(ctx, input)
}
