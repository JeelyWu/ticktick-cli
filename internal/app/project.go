package app

import (
	"context"
	"sort"

	"github.com/jeely/ticktick-cli/internal/domain"
)

type ProjectTokenSource interface {
	AccessToken(context.Context) (string, error)
}

type ProjectAPI interface {
	ListProjects(context.Context, string) ([]domain.Project, error)
	CreateProject(context.Context, string, string, string, string) (domain.Project, error)
	UpdateProject(context.Context, string, string, string, string, string) (domain.Project, error)
	DeleteProject(context.Context, string, string) error
}

type ProjectApp struct {
	Auth   ProjectTokenSource
	Client ProjectAPI
}

func (a ProjectApp) List(ctx context.Context) ([]domain.Project, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	projects, err := a.Client.ListProjects(ctx, token)
	if err != nil {
		return nil, err
	}
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].SortOrder < projects[j].SortOrder
	})
	return projects, nil
}

func (a ProjectApp) Get(ctx context.Context, ref string) (domain.Project, error) {
	projects, err := a.List(ctx)
	if err != nil {
		return domain.Project{}, err
	}
	return ResolveProject(ref, projects)
}

func (a ProjectApp) Create(ctx context.Context, name, color, kind string) (domain.Project, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return domain.Project{}, err
	}
	if kind == "" {
		kind = "TASK"
	}
	return a.Client.CreateProject(ctx, token, name, color, kind)
}

func (a ProjectApp) Update(ctx context.Context, ref, name, color, kind string) (domain.Project, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return domain.Project{}, err
	}
	projects, err := a.Client.ListProjects(ctx, token)
	if err != nil {
		return domain.Project{}, err
	}
	project, err := ResolveProject(ref, projects)
	if err != nil {
		return domain.Project{}, err
	}
	if name == "" {
		name = project.Name
	}
	if color == "" {
		color = project.Color
	}
	if kind == "" {
		kind = project.Kind
	}
	return a.Client.UpdateProject(ctx, token, project.ID, name, color, kind)
}

func (a ProjectApp) Remove(ctx context.Context, ref string) error {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return err
	}
	projects, err := a.Client.ListProjects(ctx, token)
	if err != nil {
		return err
	}
	project, err := ResolveProject(ref, projects)
	if err != nil {
		return err
	}
	return a.Client.DeleteProject(ctx, token, project.ID)
}

func ResolveProject(ref string, projects []domain.Project) (domain.Project, error) {
	matches := make([]domain.Project, 0, 1)
	names := make([]string, 0, 1)
	for _, project := range projects {
		if project.ID == ref || project.Name == ref {
			matches = append(matches, project)
			names = append(names, project.Name)
		}
	}
	switch len(matches) {
	case 0:
		return domain.Project{}, &domain.ReferenceError{Kind: "project", Reference: ref}
	case 1:
		return matches[0], nil
	default:
		return domain.Project{}, &domain.ReferenceError{Kind: "project", Reference: ref, Matches: names}
	}
}
