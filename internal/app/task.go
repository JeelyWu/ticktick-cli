package app

import (
	"context"
	"sort"
	"time"

	"github.com/jeely/ticktick-cli/internal/config"
	"github.com/jeely/ticktick-cli/internal/domain"
)

type TaskAPI interface {
	ListProjects(context.Context, string) ([]domain.Project, error)
	FilterTasks(context.Context, string, domain.TaskFilter) ([]domain.Task, error)
	GetProjectData(context.Context, string, string) (domain.Project, []domain.Task, error)
	CreateTask(context.Context, string, domain.CreateTaskPayload) (domain.Task, error)
	UpdateTask(context.Context, string, domain.Task) (domain.Task, error)
	CompleteTask(context.Context, string, string, string) error
	DeleteTask(context.Context, string, string, string) error
	MoveTask(context.Context, string, string, string, string) error
}

type TaskApp struct {
	Auth        ProjectTokenSource
	Client      TaskAPI
	ConfigStore *config.Store
	Now         func() time.Time
}

type ListTasksInput struct {
	Project    string
	Tags       []string
	Statuses   []domain.TaskStatus
	Priorities []domain.Priority
	From       string
	To         string
}

func (a TaskApp) List(ctx context.Context, in ListTasksInput) ([]domain.Task, map[string]string, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return nil, nil, err
	}
	projects, err := a.Client.ListProjects(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	projectNames := make(map[string]string, len(projects))
	filter := domain.TaskFilter{
		Tags:       in.Tags,
		Statuses:   in.Statuses,
		Priorities: in.Priorities,
	}
	if in.From != "" {
		start, err := parseListTime(in.From)
		if err != nil {
			return nil, nil, err
		}
		filter.Start = &start
	}
	if in.To != "" {
		end, err := parseListTime(in.To)
		if err != nil {
			return nil, nil, err
		}
		filter.End = &end
	}
	if in.Project != "" {
		project, err := ResolveProject(in.Project, projects)
		if err != nil {
			return nil, nil, err
		}
		filter.ProjectIDs = []string{project.ID}
	}
	for _, project := range projects {
		projectNames[project.ID] = project.Name
	}
	tasks, err := a.Client.FilterTasks(ctx, token, filter)
	if err != nil {
		return nil, nil, err
	}
	sortTasks(tasks)
	return tasks, projectNames, nil
}

func (a TaskApp) Get(ctx context.Context, ref string, projectRef string) (domain.Task, map[string]string, error) {
	input := ListTasksInput{
		Project:  projectRef,
		Statuses: []domain.TaskStatus{domain.StatusOpen},
	}
	tasks, names, err := a.List(ctx, input)
	if err != nil {
		return domain.Task{}, nil, err
	}
	task, err := resolveTaskReference(ref, tasks)
	if err != nil {
		return domain.Task{}, nil, err
	}
	return task, names, nil
}

func (a TaskApp) Today(ctx context.Context) ([]domain.Task, map[string]string, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return nil, nil, err
	}
	projects, err := a.Client.ListProjects(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now()
	if a.Now != nil {
		now = a.Now()
	}
	projectNames := make(map[string]string, len(projects))
	var tasks []domain.Task
	for _, project := range projects {
		projectNames[project.ID] = project.Name
		_, projectTasks, err := a.Client.GetProjectData(ctx, token, project.ID)
		if err != nil {
			return nil, nil, err
		}
		for _, task := range projectTasks {
			if taskIsDueTodayOrOverdue(task, now) {
				tasks = append(tasks, task)
			}
		}
	}
	sortTasks(tasks)
	return tasks, projectNames, nil
}

func (a TaskApp) Inbox(ctx context.Context) ([]domain.Task, map[string]string, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return nil, nil, err
	}

	ref := "Inbox"
	if a.ConfigStore != nil {
		cfg, err := a.ConfigStore.Load()
		if err != nil {
			return nil, nil, err
		}
		if cfg.Task.InboxProjectID != "" {
			ref = cfg.Task.InboxProjectID
		}
	}

	projects, err := a.Client.ListProjects(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	project, err := ResolveProject(ref, projects)
	if err != nil {
		return nil, nil, err
	}
	_, tasks, err := a.Client.GetProjectData(ctx, token, project.ID)
	if err != nil {
		return nil, nil, err
	}
	sortTasks(tasks)
	return tasks, map[string]string{project.ID: project.Name}, nil
}

func resolveTaskReference(ref string, tasks []domain.Task) (domain.Task, error) {
	matches := make([]domain.Task, 0, 1)
	names := make([]string, 0, 1)
	for _, task := range tasks {
		if task.ID == ref || task.Title == ref {
			matches = append(matches, task)
			names = append(names, task.Title)
		}
	}
	switch len(matches) {
	case 0:
		return domain.Task{}, &domain.ReferenceError{Kind: "task", Reference: ref}
	case 1:
		return matches[0], nil
	default:
		return domain.Task{}, &domain.ReferenceError{Kind: "task", Reference: ref, Matches: names}
	}
}

func taskIsDueTodayOrOverdue(task domain.Task, now time.Time) bool {
	if task.Status != domain.StatusOpen || task.DueDate == nil {
		return false
	}
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfToday := startOfToday.Add(24 * time.Hour)
	return task.DueDate.Before(endOfToday)
}

func sortTasks(tasks []domain.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].DueDate == nil {
			return false
		}
		if tasks[j].DueDate == nil {
			return true
		}
		return tasks[i].DueDate.Before(*tasks[j].DueDate)
	})
}

func parseListTime(value string) (time.Time, error) {
	if parsed, err := time.ParseInLocation("2006-01-02", value, time.Local); err == nil {
		return parsed, nil
	}
	return time.Parse(time.RFC3339, value)
}

func (a TaskApp) Add(ctx context.Context, in domain.CreateTaskInput) (domain.Task, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return domain.Task{}, err
	}
	projects, err := a.Client.ListProjects(ctx, token)
	if err != nil {
		return domain.Task{}, err
	}
	project, err := ResolveProject(in.ProjectRef, projects)
	if err != nil {
		return domain.Task{}, err
	}
	payload := domain.CreateTaskPayload{
		ProjectID:   project.ID,
		Title:       in.Title,
		Content:     in.Content,
		Description: in.Description,
		AllDay:      in.AllDay,
		Priority:    in.Priority,
	}
	loc := time.Local
	if in.StartRaw != "" {
		start, err := domain.ParseUserTime(in.StartRaw, loc)
		if err != nil {
			return domain.Task{}, err
		}
		payload.StartDate = &start
	}
	if in.DueRaw != "" {
		due, err := domain.ParseUserTime(in.DueRaw, loc)
		if err != nil {
			return domain.Task{}, err
		}
		payload.DueDate = &due
	}
	return a.Client.CreateTask(ctx, token, payload)
}

func (a TaskApp) Update(ctx context.Context, in domain.UpdateTaskInput) (domain.Task, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return domain.Task{}, err
	}
	task, _, err := a.Get(ctx, in.Reference, in.ProjectRef)
	if err != nil {
		return domain.Task{}, err
	}
	if in.Title != "" {
		task.Title = in.Title
	}
	if in.Content != "" {
		task.Content = in.Content
	}
	if in.Description != "" {
		task.Description = in.Description
	}
	if in.AllDay != nil {
		task.IsAllDay = *in.AllDay
	}
	if in.Priority != nil {
		task.Priority = *in.Priority
	}
	loc := time.Local
	if in.StartRaw != "" {
		start, err := domain.ParseUserTime(in.StartRaw, loc)
		if err != nil {
			return domain.Task{}, err
		}
		task.StartDate = &start
	}
	if in.DueRaw != "" {
		due, err := domain.ParseUserTime(in.DueRaw, loc)
		if err != nil {
			return domain.Task{}, err
		}
		task.DueDate = &due
	}
	return a.Client.UpdateTask(ctx, token, task)
}

func (a TaskApp) Done(ctx context.Context, ref string, projectRef string) error {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return err
	}
	task, _, err := a.Get(ctx, ref, projectRef)
	if err != nil {
		return err
	}
	return a.Client.CompleteTask(ctx, token, task.ProjectID, task.ID)
}

func (a TaskApp) Remove(ctx context.Context, ref string, projectRef string) error {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return err
	}
	task, _, err := a.Get(ctx, ref, projectRef)
	if err != nil {
		return err
	}
	return a.Client.DeleteTask(ctx, token, task.ProjectID, task.ID)
}

func (a TaskApp) Move(ctx context.Context, in domain.MoveTaskInput) error {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return err
	}
	task, _, err := a.Get(ctx, in.Reference, in.FromProjectRef)
	if err != nil {
		return err
	}
	projects, err := a.Client.ListProjects(ctx, token)
	if err != nil {
		return err
	}
	destination, err := ResolveProject(in.ToProjectRef, projects)
	if err != nil {
		return err
	}
	return a.Client.MoveTask(ctx, token, task.ProjectID, destination.ID, task.ID)
}
