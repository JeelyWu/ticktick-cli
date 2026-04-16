package main

import (
	"fmt"
	"os"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/jeely/ticktick-cli/internal/auth"
	"github.com/jeely/ticktick-cli/internal/cli"
	"github.com/jeely/ticktick-cli/internal/config"
	"github.com/jeely/ticktick-cli/internal/endpoint"
	"github.com/jeely/ticktick-cli/internal/ticktick"
	"github.com/pkg/browser"
)

type browserOpener struct{}

func (browserOpener) OpenURL(url string) error {
	return browser.OpenURL(url)
}

var version = "dev"

func loadConfigStore() (*config.Store, config.Config, error) {
	configPath, err := config.DefaultPath()
	if err != nil {
		return nil, config.Config{}, err
	}
	store := config.NewStore(configPath)
	cfg, err := store.Load()
	if err != nil {
		return nil, config.Config{}, err
	}
	return store, cfg, nil
}

func resolveRegion() (string, error) {
	_, cfg, err := loadConfigStore()
	if err != nil {
		return "", err
	}
	return cfg.Service.Region, nil
}

func buildRuntime(streams cli.Streams) (*config.Store, auth.Service, *ticktick.Client, error) {
	store, cfg, err := loadConfigStore()
	if err != nil {
		return nil, auth.Service{}, nil, err
	}
	endpoints, err := endpoint.ForRegion(cfg.Service.Region)
	if err != nil {
		return nil, auth.Service{}, nil, err
	}
	return store, auth.Service{
		AuthorizeURL: endpoints.AuthorizeURL,
		ClientID:     cfg.OAuth.ClientID,
		Exchanger: auth.Exchanger{
			TokenURL: endpoints.TokenURL,
		},
		Store:   auth.KeyringStore{},
		Browser: browserOpener{},
		In:      streams.In,
		Out:     streams.Out,
	}, ticktick.New(endpoints.APIBaseURL, nil), nil
}

func main() {
	streams := cli.Streams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	cmd := cli.NewRootCommand(cli.RootOptions{
		Version:        version,
		Streams:        streams,
		RegionResolver: resolveRegion,
		LoginAuthResolver: func() (*app.AuthApp, error) {
			store, service, _, err := buildRuntime(streams)
			if err != nil {
				return nil, err
			}
			return &app.AuthApp{
				ConfigStore: store,
				Service:     service,
			}, nil
		},
		AuthServiceResolver: func() (app.AuthService, error) {
			return auth.Service{
				Store:   auth.KeyringStore{},
				Browser: browserOpener{},
				In:      streams.In,
				Out:     streams.Out,
			}, nil
		},
		ProjectResolver: func() (*app.ProjectApp, error) {
			_, authService, api, err := buildRuntime(streams)
			if err != nil {
				return nil, err
			}
			return &app.ProjectApp{
				Auth:   authService,
				Client: api,
			}, nil
		},
		TaskResolver: func() (*app.TaskApp, error) {
			store, authService, api, err := buildRuntime(streams)
			if err != nil {
				return nil, err
			}

			return &app.TaskApp{
				Auth:        authService,
				Client:      api,
				ConfigStore: store,
			}, nil
		},
		QuickResolver: func() (*app.QuickAddApp, error) {
			store, authService, api, err := buildRuntime(streams)
			if err != nil {
				return nil, err
			}

			taskApp := &app.TaskApp{
				Auth:        authService,
				Client:      api,
				ConfigStore: store,
			}
			return &app.QuickAddApp{
				TaskApp:     taskApp,
				ConfigStore: store,
			}, nil
		},
		ConfigResolver: func() (*app.ConfigApp, error) {
			configPath, err := config.DefaultPath()
			if err != nil {
				return nil, err
			}
			return &app.ConfigApp{
				Store: config.NewStore(configPath),
			}, nil
		},
	})

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(streams.ErrOut, err)
		os.Exit(1)
	}
}
