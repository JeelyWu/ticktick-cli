package cli

import (
	"fmt"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/spf13/cobra"
)

type AuthResolver func() (*app.AuthApp, error)

type RootOptions struct {
	Version      string
	Streams      Streams
	AuthResolver AuthResolver
}

func NewRootCommand(opts RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "tick",
		Short:         "TickTick CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	cmd.SetIn(opts.Streams.In)
	cmd.SetOut(opts.Streams.Out)
	cmd.SetErr(opts.Streams.ErrOut)
	cmd.AddCommand(newVersionCommand(opts))
	if opts.AuthResolver != nil {
		cmd.AddCommand(NewAuthCommand(opts.AuthResolver, opts.Streams))
	}
	return cmd
}

func newVersionCommand(opts RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the tick version",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(opts.Streams.Out, opts.Version)
			return err
		},
	}
}
