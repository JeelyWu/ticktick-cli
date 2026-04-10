package cli

import (
	"fmt"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/spf13/cobra"
)

type AuthResolver func() (*app.AuthApp, error)
type AuthServiceResolver func() (app.AuthService, error)
type ProjectResolver func() (*app.ProjectApp, error)
type TaskResolver func() (*app.TaskApp, error)

type RootOptions struct {
	Version             string
	Streams             Streams
	LoginAuthResolver   AuthResolver
	AuthServiceResolver AuthServiceResolver
	ProjectResolver     ProjectResolver
	TaskResolver        TaskResolver
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
	if opts.LoginAuthResolver != nil || opts.AuthServiceResolver != nil {
		cmd.AddCommand(NewAuthCommand(opts.LoginAuthResolver, opts.AuthServiceResolver, opts.Streams))
	}
	if opts.ProjectResolver != nil {
		cmd.AddCommand(NewProjectCommand(opts.ProjectResolver, opts.Streams))
	}
	if opts.TaskResolver != nil {
		cmd.AddCommand(NewTaskCommand(opts.TaskResolver, opts.Streams))
		cmd.AddCommand(NewTodayCommand(opts.TaskResolver, opts.Streams))
		cmd.AddCommand(NewInboxCommand(opts.TaskResolver, opts.Streams))
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
