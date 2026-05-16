package app

import (
	"context"
	"fmt"

	"github.com/jeelywu/ticktick-cli/internal/config"
	"github.com/jeelywu/ticktick-cli/internal/endpoint"
	"gopkg.in/yaml.v3"
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
	case "output":
		return cfg.Output, nil
	case "region":
		return cfg.Region, nil
	case "default_project":
		return cfg.DefaultProject, nil
	case "client_id":
		return cfg.ClientID, nil
	case "client_secret":
		return cfg.ClientSecret, nil
	case "redirect_url":
		return cfg.RedirectURL, nil
	default:
		return "", fmt.Errorf("unsupported config key %q", key)
	}
}

func (a ConfigApp) List(ctx context.Context) (string, error) {
	cfg, err := a.Store.Load()
	if err != nil {
		return "", err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a ConfigApp) Set(ctx context.Context, key, value string) error {
	cfg, err := a.Store.Load()
	if err != nil {
		return err
	}
	switch key {
	case "output":
		if value != "table" && value != "json" {
			return fmt.Errorf("unsupported value %q for output", value)
		}
		cfg.Output = value
	case "region":
		if _, err := endpoint.ForRegion(value); err != nil {
			return err
		}
		cfg.Region = value
	case "default_project":
		cfg.DefaultProject = value
	case "client_id":
		cfg.ClientID = value
	case "client_secret":
		cfg.ClientSecret = value
	case "redirect_url":
		cfg.RedirectURL = value
	default:
		return fmt.Errorf("unsupported config key %q", key)
	}
	return a.Store.Save(cfg)
}
