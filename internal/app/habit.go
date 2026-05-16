package app

import (
	"context"
	"sort"
	"time"

	"github.com/jeelywu/ticktick-cli/internal/domain"
)

type HabitAPI interface {
	ListHabits(context.Context, string) ([]domain.Habit, error)
	GetHabit(context.Context, string, string) (domain.Habit, error)
	CreateHabit(context.Context, string, domain.CreateHabitPayload) (domain.Habit, error)
	UpdateHabit(context.Context, string, string, domain.CreateHabitPayload) (domain.Habit, error)
	CheckinHabit(context.Context, string, string, int, float64, float64) error
	ListCheckins(context.Context, string, []string, int, int) ([]domain.HabitCheckin, error)
}

type HabitApp struct {
	Auth ProjectTokenSource
	Client HabitAPI
}

func (a HabitApp) List(ctx context.Context) ([]domain.Habit, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	habits, err := a.Client.ListHabits(ctx, token)
	if err != nil {
		return nil, err
	}
	sort.Slice(habits, func(i, j int) bool {
		return habits[i].Name < habits[j].Name
	})
	return habits, nil
}

func (a HabitApp) Get(ctx context.Context, ref string) (domain.Habit, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return domain.Habit{}, err
	}
	if looksLikeID(ref) {
		return a.Client.GetHabit(ctx, token, ref)
	}
	habits, err := a.List(ctx)
	if err != nil {
		return domain.Habit{}, err
	}
	habit, err := ResolveHabit(ref, habits)
	if err == nil {
		return habit, nil
	}
	// Fallback: the habit may be archived and ListHabits only returns active habits.
	// Try a direct API call in case ref is actually an ID.
	h, apiErr := a.Client.GetHabit(ctx, token, ref)
	if apiErr == nil {
		return h, nil
	}
	return domain.Habit{}, err
}

func looksLikeID(ref string) bool {
	if len(ref) != 24 {
		return false
	}
	for _, r := range ref {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}

func (a HabitApp) Add(ctx context.Context, in domain.CreateHabitInput) (domain.Habit, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return domain.Habit{}, err
	}
	payload := domain.CreateHabitPayload{
		Name:       in.Name,
		Goal:       in.Goal,
		Color:      in.Color,
		IconRes:    in.IconRes,
		RepeatRule: in.RepeatRule,
		TargetDays: in.TargetDays,
		Unit:       in.Unit,
		Step:       in.Step,
	}
	if payload.Goal == 0 {
		payload.Goal = 1
	}
	if payload.Step == 0 {
		payload.Step = 1
	}
	return a.Client.CreateHabit(ctx, token, payload)
}

func (a HabitApp) Update(ctx context.Context, in domain.UpdateHabitInput) (domain.Habit, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return domain.Habit{}, err
	}
	habit, err := a.Get(ctx, in.Reference)
	if err != nil {
		return domain.Habit{}, err
	}
	payload := domain.CreateHabitPayload{
		Name:       habit.Name,
		Goal:       habit.Goal,
		Color:      habit.Color,
		IconRes:    habit.IconRes,
		RepeatRule: habit.RepeatRule,
		TargetDays: habit.TargetDays,
		Unit:       habit.Unit,
		Step:       habit.Step,
	}
	if in.Name != "" {
		payload.Name = in.Name
	}
	if in.Goal > 0 {
		payload.Goal = in.Goal
	}
	if in.Color != "" {
		payload.Color = in.Color
	}
	if in.IconRes != "" {
		payload.IconRes = in.IconRes
	}
	if in.RepeatRule != "" {
		payload.RepeatRule = in.RepeatRule
	}
	if in.TargetDays > 0 {
		payload.TargetDays = in.TargetDays
	}
	if in.Unit != "" {
		payload.Unit = in.Unit
	}
	if in.Step > 0 {
		payload.Step = in.Step
	}
	if in.Status != nil {
		payload.Status = int(*in.Status)
		payload.StatusSet = true
	}
	return a.Client.UpdateHabit(ctx, token, habit.ID, payload)
}

func (a HabitApp) Archive(ctx context.Context, ref string) (domain.Habit, error) {
	habit, err := a.Get(ctx, ref)
	if err != nil {
		return domain.Habit{}, err
	}
	var newStatus domain.HabitStatus
	if habit.Status == domain.HabitStatusActive {
		newStatus = domain.HabitStatusArchived
	} else {
		newStatus = domain.HabitStatusActive
	}
	return a.Update(ctx, domain.UpdateHabitInput{
		Reference: ref,
		Status:    &newStatus,
	})
}

func (a HabitApp) Checkin(ctx context.Context, ref string, value float64) error {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return err
	}
	habit, err := a.Get(ctx, ref)
	if err != nil {
		return err
	}
	if value == 0 {
		value = habit.Step
	}
	if value == 0 {
		value = 1
	}
	stamp := dateStamp(time.Now())
	return a.Client.CheckinHabit(ctx, token, habit.ID, stamp, value, habit.Goal)
}

func (a HabitApp) Log(ctx context.Context, ref string) ([]domain.HabitCheckin, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	habit, err := a.Get(ctx, ref)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	from := dateStamp(now.AddDate(0, 0, -30))
	to := dateStamp(now.AddDate(0, 0, 1))
	return a.Client.ListCheckins(ctx, token, []string{habit.ID}, from, to)
}

func dateStamp(t time.Time) int {
	return t.Year()*10000 + int(t.Month())*100 + t.Day()
}

func ResolveHabit(ref string, habits []domain.Habit) (domain.Habit, error) {
	matches := make([]domain.Habit, 0, 1)
	names := make([]string, 0, 1)
	for _, habit := range habits {
		if habit.ID == ref || habit.Name == ref {
			matches = append(matches, habit)
			names = append(names, habit.Name)
		}
	}
	switch len(matches) {
	case 0:
		return domain.Habit{}, &domain.ReferenceError{Kind: "habit", Reference: ref}
	case 1:
		return matches[0], nil
	default:
		return domain.Habit{}, &domain.ReferenceError{Kind: "habit", Reference: ref, Matches: names}
	}
}
