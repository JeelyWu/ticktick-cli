package app

import (
	"context"
	"sort"
	"time"

	"github.com/jeely/ticktick-cli/internal/config"
	"github.com/jeely/ticktick-cli/internal/domain"
)

type FocusAPI interface {
	GetFocus(context.Context, string, string) (domain.Focus, error)
	ListFocus(context.Context, string, time.Time, time.Time) ([]domain.Focus, error)
	StartFocus(context.Context, string, domain.StartFocusInput) (domain.Focus, error)
	StopFocus(context.Context, string, string) error
	ListProjects(context.Context, string) ([]domain.Project, error)
}

type FocusApp struct {
	Auth        ProjectTokenSource
	Client      FocusAPI
	ConfigStore *config.Store
	Now         func() time.Time
}

type ListFocusInput struct {
	From    string
	To      string
	Project string
}

type StartFocusAppInput struct {
	Title     string
	Content   string
	ProjectRef string
	TaskID    string
	Mode      domain.FocusMode
	StartRaw  string
}

func (a FocusApp) List(ctx context.Context, in ListFocusInput) ([]domain.Focus, map[string]string, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	if a.Now != nil {
		now = a.Now()
	}

	startDate := now.AddDate(0, 0, -7)
	if in.From != "" {
		startDate, err = parseListTime(in.From)
		if err != nil {
			return nil, nil, err
		}
	}

	endDate := now
	if in.To != "" {
		endDate, err = parseListTime(in.To)
		if err != nil {
			return nil, nil, err
		}
	}

	focuses, err := a.Client.ListFocus(ctx, token, startDate, endDate)
	if err != nil {
		return nil, nil, err
	}

	projects, err := a.Client.ListProjects(ctx, token)
	if err != nil {
		return nil, nil, err
	}

	projectNames := make(map[string]string, len(projects))
	for _, p := range projects {
		projectNames[p.ID] = p.Name
	}

	if in.Project != "" {
		project, err := ResolveProject(in.Project, projects)
		if err != nil {
			return nil, nil, err
		}
		filtered := make([]domain.Focus, 0, len(focuses))
		for _, f := range focuses {
			if f.ProjectID == project.ID {
				filtered = append(filtered, f)
			}
		}
		focuses = filtered
	}

	return focuses, projectNames, nil
}

func (a FocusApp) Get(ctx context.Context, focusID string) (domain.Focus, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return domain.Focus{}, err
	}
	return a.Client.GetFocus(ctx, token, focusID)
}

func (a FocusApp) Start(ctx context.Context, in StartFocusAppInput) (domain.Focus, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return domain.Focus{}, err
	}

	projects, err := a.Client.ListProjects(ctx, token)
	if err != nil {
		return domain.Focus{}, err
	}

	project, err := ResolveProject(in.ProjectRef, projects)
	if err != nil {
		return domain.Focus{}, err
	}

	payload := domain.StartFocusInput{
		Title:     in.Title,
		Content:   in.Content,
		ProjectID: project.ID,
		TaskID:    in.TaskID,
		Mode:      in.Mode,
	}

	if in.StartRaw != "" {
		loc := time.Local
		start, err := domain.ParseUserTime(in.StartRaw, loc)
		if err != nil {
			return domain.Focus{}, err
		}
		payload.StartDate = &start
	}

	return a.Client.StartFocus(ctx, token, payload)
}

func (a FocusApp) Stop(ctx context.Context, focusID string) error {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return err
	}
	return a.Client.StopFocus(ctx, token, focusID)
}

func (a FocusApp) ListProjects(ctx context.Context) ([]domain.Project, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	projects, err := a.Client.ListProjects(ctx, token)
	if err != nil {
		return nil, err
	}
	sortProjects(projects)
	return projects, nil
}

func sortProjects(projects []domain.Project) {
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].SortOrder < projects[j].SortOrder
	})
}
