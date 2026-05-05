package ticktick

import (
	"context"
	"net/http"
	"sort"

	"github.com/jeely/ticktick-cli/internal/domain"
)

type habitDTO struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Status        int      `json:"status"`
	Goal          int      `json:"goal"`
	Color         string   `json:"color"`
	SectionID     string   `json:"sectionId"`
	Icon          string   `json:"icon"`
	Reminders     []string `json:"reminders"`
	RepeatRule    string   `json:"repeatRule"`
	TargetDays    []int    `json:"targetDays"`
	Unit          string   `json:"unit"`
	Step          int      `json:"step"`
	TotalCheckins int      `json:"totalCheckins"`
	CurrentStreak int      `json:"currentStreak"`
	CreatedTime   string   `json:"createdTime"`
	ArchivedTime  string   `json:"archivedTime"`
}

type habitCheckinDTO struct {
	ID           string `json:"id"`
	HabitID      string `json:"habitId"`
	CheckinStamp int    `json:"checkinStamp"`
	CheckinTime  string `json:"checkinTime"`
	Status       int    `json:"status"`
	Value        int    `json:"value"`
	Goal         int    `json:"goal"`
}

type habitsResponse struct {
	Habits []habitDTO `json:"habits"`
}

func mapHabits(dto []habitDTO) []domain.Habit {
	out := make([]domain.Habit, 0, len(dto))
	for _, item := range dto {
		out = append(out, domain.Habit{
			ID:            item.ID,
			Name:          item.Name,
			Status:        domain.HabitStatus(item.Status),
			Goal:          item.Goal,
			Color:         item.Color,
			SectionID:     item.SectionID,
			Icon:          item.Icon,
			Reminders:     item.Reminders,
			RepeatRule:    item.RepeatRule,
			TargetDays:    item.TargetDays,
			Unit:          item.Unit,
			Step:          item.Step,
			TotalCheckins: item.TotalCheckins,
			CurrentStreak: item.CurrentStreak,
			CreatedTime:   parseTickTime(item.CreatedTime),
			ArchivedTime:  parseTickTime(item.ArchivedTime),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func mapCheckins(dto []habitCheckinDTO) []domain.HabitCheckin {
	out := make([]domain.HabitCheckin, 0, len(dto))
	for _, item := range dto {
		out = append(out, domain.HabitCheckin{
			ID:           item.ID,
			HabitID:      item.HabitID,
			CheckinStamp: item.CheckinStamp,
			CheckinTime:  parseTickTime(item.CheckinTime),
			Status:       item.Status,
			Value:        item.Value,
			Goal:         item.Goal,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CheckinStamp > out[j].CheckinStamp
	})
	return out
}

func (c *Client) ListHabits(ctx context.Context, token string) ([]domain.Habit, error) {
	var resp habitsResponse
	if err := c.DoJSON(ctx, http.MethodGet, "/open/v1/habits", token, nil, &resp); err != nil {
		return nil, err
	}
	return mapHabits(resp.Habits), nil
}

func (c *Client) GetHabit(ctx context.Context, token, habitID string) (domain.Habit, error) {
	var dto habitDTO
	if err := c.DoJSON(ctx, http.MethodGet, "/open/v1/habits/"+habitID, token, nil, &dto); err != nil {
		return domain.Habit{}, err
	}
	habits := mapHabits([]habitDTO{dto})
	return habits[0], nil
}

func (c *Client) CreateHabit(ctx context.Context, token string, in domain.CreateHabitPayload) (domain.Habit, error) {
	body := map[string]any{
		"name": in.Name,
		"goal": in.Goal,
	}
	if in.Color != "" {
		body["color"] = in.Color
	}
	if in.Icon != "" {
		body["icon"] = in.Icon
	}
	if in.RepeatRule != "" {
		body["repeatRule"] = in.RepeatRule
	}
	if len(in.TargetDays) > 0 {
		body["targetDays"] = in.TargetDays
	}
	if in.Unit != "" {
		body["unit"] = in.Unit
	}
	if in.Step > 0 {
		body["step"] = in.Step
	}
	var dto habitDTO
	if err := c.DoJSON(ctx, http.MethodPost, "/open/v1/habits", token, body, &dto); err != nil {
		return domain.Habit{}, err
	}
	return mapHabits([]habitDTO{dto})[0], nil
}

func (c *Client) UpdateHabit(ctx context.Context, token, habitID string, in domain.CreateHabitPayload) (domain.Habit, error) {
	body := map[string]any{
		"id":   habitID,
		"name": in.Name,
		"goal": in.Goal,
	}
	if in.Color != "" {
		body["color"] = in.Color
	}
	if in.Icon != "" {
		body["icon"] = in.Icon
	}
	if in.RepeatRule != "" {
		body["repeatRule"] = in.RepeatRule
	}
	if len(in.TargetDays) > 0 {
		body["targetDays"] = in.TargetDays
	}
	if in.Unit != "" {
		body["unit"] = in.Unit
	}
	if in.Step > 0 {
		body["step"] = in.Step
	}
	var dto habitDTO
	if err := c.DoJSON(ctx, http.MethodPost, "/open/v1/habits/"+habitID, token, body, &dto); err != nil {
		return domain.Habit{}, err
	}
	return mapHabits([]habitDTO{dto})[0], nil
}

func (c *Client) DeleteHabit(ctx context.Context, token, habitID string) error {
	return c.DoJSON(ctx, http.MethodDelete, "/open/v1/habits/"+habitID, token, nil, nil)
}

func (c *Client) CheckinHabit(ctx context.Context, token, habitID string, value int) error {
	body := map[string]any{
		"value": value,
	}
	return c.DoJSON(ctx, http.MethodPost, "/open/v1/habits/"+habitID+"/checkins", token, body, nil)
}

func (c *Client) ListCheckins(ctx context.Context, token, habitID string) ([]domain.HabitCheckin, error) {
	var dto []habitCheckinDTO
	if err := c.DoJSON(ctx, http.MethodGet, "/open/v1/habits/"+habitID+"/checkins", token, nil, &dto); err != nil {
		return nil, err
	}
	return mapCheckins(dto), nil
}
