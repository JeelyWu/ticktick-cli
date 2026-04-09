package main

import (
	"fmt"
	"os"

	"github.com/jeely/ticktick-cli/internal/cli"
)

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
	})

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(streams.ErrOut, err)
		os.Exit(1)
	}
}
