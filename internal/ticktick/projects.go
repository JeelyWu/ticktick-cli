package ticktick

import (
	"context"
	"net/http"

	"github.com/jeely/ticktick-cli/internal/domain"
)

type projectDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	SortOrder  int64  `json:"sortOrder"`
	Closed     bool   `json:"closed"`
	GroupID    string `json:"groupId"`
	ViewMode   string `json:"viewMode"`
	Permission string `json:"permission"`
	Kind       string `json:"kind"`
}

func (c *Client) ListProjects(ctx context.Context, token string) ([]domain.Project, error) {
	var dto []projectDTO
	if err := c.DoJSON(ctx, http.MethodGet, "/open/v1/project", token, nil, &dto); err != nil {
		return nil, err
	}
	projects := make([]domain.Project, 0, len(dto))
	for _, item := range dto {
		projects = append(projects, domain.Project{
			ID:         item.ID,
			Name:       item.Name,
			Color:      item.Color,
			SortOrder:  item.SortOrder,
			Closed:     item.Closed,
			GroupID:    item.GroupID,
			ViewMode:   item.ViewMode,
			Permission: item.Permission,
			Kind:       item.Kind,
		})
	}
	return projects, nil
}

func (c *Client) CreateProject(ctx context.Context, token, name, color, kind string) (domain.Project, error) {
	body := map[string]any{
		"name":  name,
		"color": color,
		"kind":  kind,
	}
	var dto projectDTO
	if err := c.DoJSON(ctx, http.MethodPost, "/open/v1/project", token, body, &dto); err != nil {
		return domain.Project{}, err
	}
	return domain.Project{
		ID:         dto.ID,
		Name:       dto.Name,
		Color:      dto.Color,
		SortOrder:  dto.SortOrder,
		Closed:     dto.Closed,
		GroupID:    dto.GroupID,
		ViewMode:   dto.ViewMode,
		Permission: dto.Permission,
		Kind:       dto.Kind,
	}, nil
}

func (c *Client) UpdateProject(ctx context.Context, token, projectID, name, color, kind string) (domain.Project, error) {
	body := map[string]any{
		"name":  name,
		"color": color,
		"kind":  kind,
	}
	var dto projectDTO
	if err := c.DoJSON(ctx, http.MethodPost, "/open/v1/project/"+projectID, token, body, &dto); err != nil {
		return domain.Project{}, err
	}
	return domain.Project{
		ID:         dto.ID,
		Name:       dto.Name,
		Color:      dto.Color,
		SortOrder:  dto.SortOrder,
		Closed:     dto.Closed,
		GroupID:    dto.GroupID,
		ViewMode:   dto.ViewMode,
		Permission: dto.Permission,
		Kind:       dto.Kind,
	}, nil
}

func (c *Client) DeleteProject(ctx context.Context, token, projectID string) error {
	return c.DoJSON(ctx, http.MethodDelete, "/open/v1/project/"+projectID, token, nil, nil)
}
