package main

import (
	"fmt"
	"os"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/jeely/ticktick-cli/internal/auth"
	"github.com/jeely/ticktick-cli/internal/cli"
	"github.com/jeely/ticktick-cli/internal/config"
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

	configPath, err := config.DefaultPath()
	if err != nil {
		_, _ = fmt.Fprintln(streams.ErrOut, err)
		os.Exit(1)
	}

	authApp := &app.AuthApp{
		ConfigStore: config.NewStore(configPath),
		Service: auth.Service{
			Store:   auth.KeyringStore{},
			Browser: browserOpener{},
			In:      streams.In,
			Out:     streams.Out,
		},
	}

	cmd := cli.NewRootCommand(cli.RootOptions{
		Version: version,
		Streams: streams,
		AuthApp: authApp,
	})

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(streams.ErrOut, err)
		os.Exit(1)
	}
}
