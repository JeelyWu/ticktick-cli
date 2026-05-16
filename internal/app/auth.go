package app

import (
	"context"
	"errors"

	"github.com/jeelywu/ticktick-cli/internal/auth"
	"github.com/jeelywu/ticktick-cli/internal/config"
)

type AuthService interface {
	Login(context.Context, auth.LoginInput) (auth.Token, error)
	Status(context.Context) (auth.Status, error)
	Logout(context.Context) error
}

type AuthServiceFactory func(region string) (AuthService, error)

type AuthApp struct {
	ConfigStore      *config.Store
	Service          AuthService
	ServiceForRegion AuthServiceFactory
}

type LoginInput struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Region       string
}

func (a AuthApp) Login(ctx context.Context, in LoginInput) error {
	cfg := config.Default()
	if a.ConfigStore != nil {
		needsDefaults := in.ClientID == "" || in.ClientSecret == ""
		loaded, err := a.ConfigStore.Load()
		if err != nil {
			if needsDefaults {
				return err
			}
		} else {
			cfg = loaded
		}
	}
	if in.ClientID == "" {
		in.ClientID = cfg.ClientID
	}
	if in.ClientSecret == "" {
		in.ClientSecret = cfg.ClientSecret
	}
	if in.Region == "" {
		in.Region = cfg.Region
	}
	if in.RedirectURL == "" {
		if cfg.RedirectURL != "" {
			in.RedirectURL = cfg.RedirectURL
		} else {
			in.RedirectURL = "http://localhost:8080/callback"
		}
	}
	if in.ClientID == "" || in.ClientSecret == "" || in.RedirectURL == "" {
		return errors.New("login requires client-id, client-secret, and redirect-url")
	}
	service := a.Service
	if a.ServiceForRegion != nil {
		var err error
		service, err = a.ServiceForRegion(in.Region)
		if err != nil {
			return err
		}
	}
	if service == nil {
		return errors.New("auth service is unavailable")
	}
	if a.ConfigStore != nil {
		cfg.ClientID = in.ClientID
		cfg.ClientSecret = in.ClientSecret
		cfg.Region = in.Region
		cfg.RedirectURL = in.RedirectURL
		if err := a.ConfigStore.Save(cfg); err != nil {
			return err
		}
	}
	_, err := service.Login(ctx, auth.LoginInput{
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
