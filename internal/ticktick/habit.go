package ticktick

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/jeelywu/ticktick-cli/internal/domain"
)

type habitDTO struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Status        int      `json:"status"`
	Goal          float64  `json:"goal"`
	Color         string   `json:"color"`
	SectionID     string   `json:"sectionId"`
	IconRes       string   `json:"iconRes"`
	Reminders     []string `json:"reminders"`
	RepeatRule    string   `json:"repeatRule"`
	TargetDays    int      `json:"targetDays"`
	Unit          string   `json:"unit"`
	Step          float64  `json:"step"`
	TotalCheckins int      `json:"totalCheckins"`
	CurrentStreak int      `json:"currentStreak"`
	CreatedTime   string   `json:"createdTime"`
	ArchivedTime  string   `json:"archivedTime"`
}

type checkinEntryDTO struct {
	Stamp int     `json:"stamp"`
	Value float64 `json:"value"`
	Goal  float64 `json:"goal"`
}

type habitCheckinGroupDTO struct {
	HabitID  string            `json:"habitId"`
	Year     int               `json:"year"`
	Checkins []checkinEntryDTO `json:"checkins"`
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
			IconRes:       item.IconRes,
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

func mapCheckins(group habitCheckinGroupDTO) []domain.HabitCheckin {
	out := make([]domain.HabitCheckin, 0, len(group.Checkins))
	for _, item := range group.Checkins {
		out = append(out, domain.HabitCheckin{
			HabitID: group.HabitID,
			Year:    group.Year,
			Stamp:   item.Stamp,
			Value:   item.Value,
			Goal:    item.Goal,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Stamp > out[j].Stamp
	})
	return out
}

func (c *Client) ListHabits(ctx context.Context, token string) ([]domain.Habit, error) {
	var resp []habitDTO
	if err := c.DoJSON(ctx, http.MethodGet, "/open/v1/habit", token, nil, &resp); err != nil {
		return nil, err
	}
	return mapHabits(resp), nil
}

func (c *Client) GetHabit(ctx context.Context, token, habitID string) (domain.Habit, error) {
	var dto habitDTO
	if err := c.DoJSON(ctx, http.MethodGet, "/open/v1/habit/"+habitID, token, nil, &dto); err != nil {
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
	if in.IconRes != "" {
		body["iconRes"] = in.IconRes
	}
	if in.RepeatRule != "" {
		body["repeatRule"] = in.RepeatRule
	}
	if in.TargetDays > 0 {
		body["targetDays"] = in.TargetDays
	}
	if in.Unit != "" {
		body["unit"] = in.Unit
	}
	if in.Step > 0 {
		body["step"] = in.Step
	}
	var dto habitDTO
	if err := c.DoJSON(ctx, http.MethodPost, "/open/v1/habit", token, body, &dto); err != nil {
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
	if in.IconRes != "" {
		body["iconRes"] = in.IconRes
	}
	if in.RepeatRule != "" {
		body["repeatRule"] = in.RepeatRule
	}
	if in.TargetDays > 0 {
		body["targetDays"] = in.TargetDays
	}
	if in.Unit != "" {
		body["unit"] = in.Unit
	}
	if in.Step > 0 {
		body["step"] = in.Step
	}
	if in.StatusSet {
		body["status"] = in.Status
	}
	var dto habitDTO
	if err := c.DoJSON(ctx, http.MethodPost, "/open/v1/habit/"+habitID, token, body, &dto); err != nil {
		return domain.Habit{}, err
	}
	return mapHabits([]habitDTO{dto})[0], nil
}

func (c *Client) CheckinHabit(ctx context.Context, token, habitID string, stamp int, value, goal float64) error {
	body := map[string]any{
		"stamp": stamp,
		"value": value,
	}
	if goal > 0 {
		body["goal"] = goal
	}
	return c.DoJSON(ctx, http.MethodPost, "/open/v1/habit/"+habitID+"/checkin", token, body, nil)
}

func (c *Client) ListCheckins(ctx context.Context, token string, habitIDs []string, from, to int) ([]domain.HabitCheckin, error) {
	ids := ""
	for i, id := range habitIDs {
		if i > 0 {
			ids += ","
		}
		ids += id
	}
	path := fmt.Sprintf("/open/v1/habit/checkins?habitIds=%s&from=%d&to=%d", ids, from, to)
	var resp []habitCheckinGroupDTO
	if err := c.DoJSON(ctx, http.MethodGet, path, token, nil, &resp); err != nil {
		return nil, err
	}

	out := make([]domain.HabitCheckin, 0)
	for _, group := range resp {
		out = append(out, mapCheckins(group)...)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Stamp > out[j].Stamp
	})
	return out, nil
}
