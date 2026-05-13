package ticktick

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jeely/ticktick-cli/internal/domain"
)

type focusDTO struct {
	ID            string   `json:"id"`
	Mode          int      `json:"mode"`
	Status        int      `json:"status"`
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	ProjectID     string   `json:"projectId"`
	TaskID        string   `json:"taskId"`
	StartDate     string   `json:"startDate"`
	EndDate       string   `json:"endDate"`
	TimeZone      string   `json:"timezone"`
	AbandonReason string   `json:"abandonReason"`
	Tags          []string `json:"tags"`
	Creators      []string `json:"creators"`
	SortOrder     int      `json:"sortOrder"`
}

type focusListResponse struct {
	Focuses []focusDTO `json:"focuses"`
}

func mapFocus(dto focusDTO) domain.Focus {
	return domain.Focus{
		ID:            dto.ID,
		Mode:          domain.FocusMode(dto.Mode),
		Status:        domain.FocusStatus(dto.Status),
		Title:         dto.Title,
		Content:       dto.Content,
		ProjectID:     dto.ProjectID,
		TaskID:        dto.TaskID,
		StartDate:     parseTickTime(dto.StartDate),
		EndDate:       parseTickTime(dto.EndDate),
		TimeZone:      dto.TimeZone,
		AbandonReason: dto.AbandonReason,
		Tags:          dto.Tags,
		Creators:      dto.Creators,
		SortOrder:     dto.SortOrder,
	}
}

func (c *Client) GetFocus(ctx context.Context, token, focusID string) (domain.Focus, error) {
	var dto focusDTO
	if err := c.DoJSON(ctx, http.MethodGet, "/open/v1/focus/"+focusID, token, nil, &dto); err != nil {
		return domain.Focus{}, err
	}
	return mapFocus(dto), nil
}

func (c *Client) ListFocus(ctx context.Context, token string, startDate, endDate time.Time) ([]domain.Focus, error) {
	path := fmt.Sprintf("/open/v1/focus?from=%s&to=%s",
		startDate.Format(time.RFC3339),
		endDate.Format(time.RFC3339))
	var resp focusListResponse
	if err := c.DoJSON(ctx, http.MethodGet, path, token, nil, &resp); err != nil {
		return nil, err
	}
	out := make([]domain.Focus, 0, len(resp.Focuses))
	for _, dto := range resp.Focuses {
		out = append(out, mapFocus(dto))
	}
	return out, nil
}

func (c *Client) StartFocus(ctx context.Context, token string, in domain.StartFocusInput) (domain.Focus, error) {
	body := map[string]any{
		"title":     in.Title,
		"content":   in.Content,
		"mode":      int(in.Mode),
		"projectId": in.ProjectID,
		"taskId":    in.TaskID,
	}
	if in.StartDate != nil {
		body["startDate"] = in.StartDate.Format(time.RFC3339)
	}
	var dto focusDTO
	if err := c.DoJSON(ctx, http.MethodPost, "/open/v1/focus", token, body, &dto); err != nil {
		return domain.Focus{}, err
	}
	return mapFocus(dto), nil
}

func (c *Client) StopFocus(ctx context.Context, token, focusID string) error {
	return c.DoJSON(ctx, http.MethodPost, "/open/v1/focus/"+focusID, token, nil, nil)
}
