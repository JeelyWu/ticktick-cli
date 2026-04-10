package main

import (
	"fmt"
	"os"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/jeely/ticktick-cli/internal/auth"
	"github.com/jeely/ticktick-cli/internal/cli"
	"github.com/jeely/ticktick-cli/internal/config"
	"github.com/jeely/ticktick-cli/internal/ticktick"
	"github.com/pkg/browser"
)

type browserOpener struct{}

func (browserOpener) OpenURL(url string) error {
	return browser.OpenURL(url)
}

var version = "dev"

func main() {
	streams := cli.Streams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	cmd := cli.NewRootCommand(cli.RootOptions{
		Version: version,
		Streams: streams,
		LoginAuthResolver: func() (*app.AuthApp, error) {
			service := auth.Service{
				Store:   auth.KeyringStore{},
				Browser: browserOpener{},
				In:      streams.In,
				Out:     streams.Out,
			}

			configPath, err := config.DefaultPath()
			if err != nil {
				return &app.AuthApp{
					Service: service,
				}, nil
			}

			return &app.AuthApp{
				ConfigStore: config.NewStore(configPath),
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
			authService := auth.Service{
				Store:   auth.KeyringStore{},
				Browser: browserOpener{},
				In:      streams.In,
				Out:     streams.Out,
			}
			return &app.ProjectApp{
				Auth:   authService,
				Client: ticktick.New("", nil),
			}, nil
		},
		TaskResolver: func() (*app.TaskApp, error) {
			authService := auth.Service{
				Store:   auth.KeyringStore{},
				Browser: browserOpener{},
				In:      streams.In,
				Out:     streams.Out,
			}

			var store *config.Store
			if configPath, err := config.DefaultPath(); err == nil {
				store = config.NewStore(configPath)
			}

			return &app.TaskApp{
				Auth:        authService,
				Client:      ticktick.New("", nil),
				ConfigStore: store,
			}, nil
		},
	})

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(streams.ErrOut, err)
		os.Exit(1)
	}
}
