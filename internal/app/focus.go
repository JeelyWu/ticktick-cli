package app

import (
	"context"
	"sort"
	"time"

	"github.com/jeelywu/ticktick-cli/internal/domain"
)

type FocusAPI interface {
	GetFocus(context.Context, string, string, int) (domain.Focus, error)
	ListFocus(context.Context, string, time.Time, time.Time, int) ([]domain.Focus, error)
	ListProjects(context.Context, string) ([]domain.Project, error)
}

type FocusApp struct {
	Auth   ProjectTokenSource
	Client FocusAPI
	Now    func() time.Time
}

type ListFocusInput struct {
	From string
	To   string
	Type int
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

	focuses, err := a.Client.ListFocus(ctx, token, startDate, endDate, in.Type)
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

	return focuses, projectNames, nil
}

func (a FocusApp) Get(ctx context.Context, focusID string, focusType int) (domain.Focus, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return domain.Focus{}, err
	}
	return a.Client.GetFocus(ctx, token, focusID, focusType)
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
