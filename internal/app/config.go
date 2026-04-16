package app

import (
	"context"
	"fmt"

	"github.com/jeely/ticktick-cli/internal/config"
	"github.com/jeely/ticktick-cli/internal/endpoint"
)

type ConfigApp struct {
	Store *config.Store
}

func (a ConfigApp) Get(ctx context.Context, key string) (string, error) {
	cfg, err := a.Store.Load()
	if err != nil {
		return "", err
	}
	switch key {
	case "output.default":
		return cfg.Output.Default, nil
	case "service.region":
		return cfg.Service.Region, nil
	case "task.default_project":
		return cfg.Task.DefaultProject, nil
	case "task.inbox_project_id":
		return cfg.Task.InboxProjectID, nil
	case "oauth.client_id":
		return cfg.OAuth.ClientID, nil
	case "oauth.redirect_url":
		return cfg.OAuth.RedirectURL, nil
	default:
		return "", fmt.Errorf("unsupported config key %q", key)
	}
}

func (a ConfigApp) Set(ctx context.Context, key, value string) error {
	cfg, err := a.Store.Load()
	if err != nil {
		return err
	}
	switch key {
	case "output.default":
		cfg.Output.Default = value
	case "service.region":
		if _, err := endpoint.ForRegion(value); err != nil {
			return err
		}
		cfg.Service.Region = value
	case "task.default_project":
		cfg.Task.DefaultProject = value
	case "task.inbox_project_id":
		cfg.Task.InboxProjectID = value
	case "oauth.client_id":
		cfg.OAuth.ClientID = value
	case "oauth.redirect_url":
		cfg.OAuth.RedirectURL = value
	default:
		return fmt.Errorf("unsupported config key %q", key)
	}
	return a.Store.Save(cfg)
}
