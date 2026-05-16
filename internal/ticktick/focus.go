package ticktick

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jeely/ticktick-cli/internal/domain"
)

type focusDTO struct {
	ID            string `json:"id"`
	Type          int    `json:"type"`
	Status        int    `json:"status"`
	Note          string `json:"note"`
	TaskID        string `json:"taskId"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	Duration      int64  `json:"duration"`
	PauseDuration int    `json:"pauseDuration"`
}

type focusListResponse struct {
	Focuses []focusDTO `json:"focuses"`
}

func (r *focusListResponse) UnmarshalJSON(data []byte) error {
	var items []focusDTO
	if err := json.Unmarshal(data, &items); err == nil {
		r.Focuses = items
		return nil
	}

	var wrapped struct {
		Focuses []focusDTO `json:"focuses"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return err
	}
	r.Focuses = wrapped.Focuses
	return nil
}

func mapFocus(dto focusDTO) domain.Focus {
	return domain.Focus{
		ID:        dto.ID,
		Mode:      domain.FocusModeFromAPIType(dto.Type),
		Status:    domain.FocusStatus(dto.Status),
		Title:     dto.Note,
		TaskID:    dto.TaskID,
		StartDate: parseTickTime(dto.StartTime),
		EndDate:   parseTickTime(dto.EndTime),
	}
}

func (c *Client) GetFocus(ctx context.Context, token, focusID string, focusType int) (domain.Focus, error) {
	path := fmt.Sprintf("/open/v1/focus/%s?type=%d", focusID, focusType)
	var dto focusDTO
	if err := c.DoJSON(ctx, http.MethodGet, path, token, nil, &dto); err != nil {
		return domain.Focus{}, err
	}
	return mapFocus(dto), nil
}

func (c *Client) ListFocus(ctx context.Context, token string, startDate, endDate time.Time, focusType int) ([]domain.Focus, error) {
	const apiTimeFormat = "2006-01-02T15:04:05-0700"
	path := fmt.Sprintf("/open/v1/focus?from=%s&to=%s&type=%d",
		startDate.Format(apiTimeFormat),
		endDate.Format(apiTimeFormat),
		focusType)
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
