package app

import (
	"context"
	"sort"

	"github.com/jeely/ticktick-cli/internal/domain"
)

type HabitAPI interface {
	ListHabits(context.Context, string) ([]domain.Habit, error)
	GetHabit(context.Context, string, string) (domain.Habit, error)
	CreateHabit(context.Context, string, domain.CreateHabitPayload) (domain.Habit, error)
	UpdateHabit(context.Context, string, string, domain.CreateHabitPayload) (domain.Habit, error)
	DeleteHabit(context.Context, string, string) error
	CheckinHabit(context.Context, string, string, int) error
	ListCheckins(context.Context, string, string) ([]domain.HabitCheckin, error)
}

type HabitApp struct {
	Auth   ProjectTokenSource
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
	habits, err := a.List(ctx)
	if err != nil {
		return domain.Habit{}, err
	}
	return ResolveHabit(ref, habits)
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
		Icon:       in.Icon,
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
		Icon:       habit.Icon,
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
	if in.Icon != "" {
		payload.Icon = in.Icon
	}
	if in.RepeatRule != "" {
		payload.RepeatRule = in.RepeatRule
	}
	if len(in.TargetDays) > 0 {
		payload.TargetDays = in.TargetDays
	}
	if in.Unit != "" {
		payload.Unit = in.Unit
	}
	if in.Step > 0 {
		payload.Step = in.Step
	}
	return a.Client.UpdateHabit(ctx, token, habit.ID, payload)
}

func (a HabitApp) Archive(ctx context.Context, ref string) (domain.Habit, error) {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return domain.Habit{}, err
	}
	habit, err := a.Get(ctx, ref)
	if err != nil {
		return domain.Habit{}, err
	}
	return a.Client.UpdateHabit(ctx, token, habit.ID, domain.CreateHabitPayload{})
}

func (a HabitApp) Remove(ctx context.Context, ref string) error {
	token, err := a.Auth.AccessToken(ctx)
	if err != nil {
		return err
	}
	habit, err := a.Get(ctx, ref)
	if err != nil {
		return err
	}
	return a.Client.DeleteHabit(ctx, token, habit.ID)
}

func (a HabitApp) Checkin(ctx context.Context, ref string, value int) error {
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
	return a.Client.CheckinHabit(ctx, token, habit.ID, value)
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
	return a.Client.ListCheckins(ctx, token, habit.ID)
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
