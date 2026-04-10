package ticktick

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/jeely/ticktick-cli/internal/domain"
)

type taskDTO struct {
	ID        string   `json:"id"`
	ProjectID string   `json:"projectId"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Desc      string   `json:"desc"`
	StartDate string   `json:"startDate"`
	DueDate   string   `json:"dueDate"`
	TimeZone  string   `json:"timeZone"`
	IsAllDay  bool     `json:"isAllDay"`
	Priority  int      `json:"priority"`
	Status    int      `json:"status"`
	Tags      []string `json:"tags"`
	Kind      string   `json:"kind"`
}

type taskFilterRequest struct {
	ProjectIDs []string `json:"projectIds,omitempty"`
	StartDate  string   `json:"startDate,omitempty"`
	EndDate    string   `json:"endDate,omitempty"`
	Priority   []int    `json:"priority,omitempty"`
	Tag        []string `json:"tag,omitempty"`
	Status     []int    `json:"status,omitempty"`
}

type projectDataDTO struct {
	Project projectDTO `json:"project"`
	Tasks   []taskDTO  `json:"tasks"`
}

func (c *Client) FilterTasks(ctx context.Context, token string, filter domain.TaskFilter) ([]domain.Task, error) {
	req := taskFilterRequest{
		ProjectIDs: filter.ProjectIDs,
		Priority:   filter.PriorityCodes(),
		Tag:        filter.Tags,
		Status:     filter.StatusCodes(),
	}
	if filter.Start != nil {
		req.StartDate = filter.Start.Format("2006-01-02T15:04:05.000-0700")
	}
	if filter.End != nil {
		req.EndDate = filter.End.Format("2006-01-02T15:04:05.000-0700")
	}

	var dto []taskDTO
	if err := c.DoJSON(ctx, http.MethodPost, "/open/v1/task/filter", token, req, &dto); err != nil {
		return nil, err
	}
	return mapTasks(dto), nil
}

func (c *Client) GetProjectData(ctx context.Context, token, projectID string) (domain.Project, []domain.Task, error) {
	var dto projectDataDTO
	if err := c.DoJSON(ctx, http.MethodGet, "/open/v1/project/"+projectID+"/data", token, nil, &dto); err != nil {
		return domain.Project{}, nil, err
	}
	return domain.Project{
		ID:         dto.Project.ID,
		Name:       dto.Project.Name,
		Color:      dto.Project.Color,
		SortOrder:  dto.Project.SortOrder,
		Closed:     dto.Project.Closed,
		GroupID:    dto.Project.GroupID,
		ViewMode:   dto.Project.ViewMode,
		Permission: dto.Project.Permission,
		Kind:       dto.Project.Kind,
	}, mapTasks(dto.Tasks), nil
}

func mapTasks(dto []taskDTO) []domain.Task {
	out := make([]domain.Task, 0, len(dto))
	for _, item := range dto {
		out = append(out, domain.Task{
			ID:          item.ID,
			ProjectID:   item.ProjectID,
			Title:       item.Title,
			Content:     item.Content,
			Description: item.Desc,
			StartDate:   parseTickTime(item.StartDate),
			DueDate:     parseTickTime(item.DueDate),
			TimeZone:    item.TimeZone,
			IsAllDay:    item.IsAllDay,
			Priority:    domain.Priority(item.Priority),
			Status:      domain.TaskStatus(item.Status),
			Tags:        item.Tags,
			Kind:        item.Kind,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Title < out[j].Title
	})
	return out
}

func parseTickTime(value string) *time.Time {
	if value == "" {
		return nil
	}
	formats := []string{
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05-0700",
	}
	for _, format := range formats {
		if parsed, err := time.Parse(format, value); err == nil {
			return &parsed
		}
	}
	return nil
}
