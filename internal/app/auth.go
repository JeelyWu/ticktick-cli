package app

import (
	"context"
	"errors"

	"github.com/jeely/ticktick-cli/internal/auth"
	"github.com/jeely/ticktick-cli/internal/config"
)

type AuthService interface {
	Login(context.Context, auth.LoginInput) (auth.Token, error)
	Status(context.Context) (auth.Status, error)
	Logout(context.Context) error
}

type AuthApp struct {
	ConfigStore *config.Store
	Service     AuthService
}

type LoginInput struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func (a AuthApp) Login(ctx context.Context, in LoginInput) error {
	cfg, err := a.ConfigStore.Load()
	if err != nil {
		return err
	}
	if in.ClientID == "" {
		in.ClientID = cfg.OAuth.ClientID
	}
	if in.RedirectURL == "" {
		in.RedirectURL = cfg.OAuth.RedirectURL
	}
	if in.ClientID == "" || in.ClientSecret == "" || in.RedirectURL == "" {
		return errors.New("login requires client-id, client-secret, and redirect-url")
	}
	cfg.OAuth.ClientID = in.ClientID
	cfg.OAuth.RedirectURL = in.RedirectURL
	if err := a.ConfigStore.Save(cfg); err != nil {
		return err
	}
	_, err = a.Service.Login(ctx, auth.LoginInput{
		ClientID:     in.ClientID,
		ClientSecret: in.ClientSecret,
		RedirectURL:  in.RedirectURL,
	})
	return err
}

func (a AuthApp) Status(ctx context.Context) (auth.Status, error) {
	return a.Service.Status(ctx)
}

func (a AuthApp) Logout(ctx context.Context) error {
	return a.Service.Logout(ctx)
}
