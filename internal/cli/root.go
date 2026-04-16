package cli

import (
	"fmt"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/spf13/cobra"
)

type AuthResolver func() (*app.AuthApp, error)
type AuthServiceResolver func() (app.AuthService, error)
type RegionResolver func() (string, error)
type ProjectResolver func() (*app.ProjectApp, error)
type TaskResolver func() (*app.TaskApp, error)

type RootOptions struct {
	Version             string
	Streams             Streams
	LoginAuthResolver   AuthResolver
	AuthServiceResolver AuthServiceResolver
	RegionResolver      RegionResolver
	ProjectResolver     ProjectResolver
	TaskResolver        TaskResolver
	QuickResolver       QuickResolver
	ConfigResolver      ConfigResolver
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
	cmd.AddCommand(newVersionCommand(opts.Version, opts.RegionResolver, opts.Streams))
	if opts.LoginAuthResolver != nil || opts.AuthServiceResolver != nil {
		cmd.AddCommand(NewAuthCommand(opts.LoginAuthResolver, opts.AuthServiceResolver, opts.RegionResolver, opts.Streams))
	}
	if opts.ProjectResolver != nil {
		cmd.AddCommand(NewProjectCommand(opts.ProjectResolver, opts.Streams))
	}
	if opts.TaskResolver != nil {
		cmd.AddCommand(NewTaskCommand(opts.TaskResolver, opts.Streams))
		cmd.AddCommand(NewTodayCommand(opts.TaskResolver, opts.Streams))
		cmd.AddCommand(NewInboxCommand(opts.TaskResolver, opts.Streams))
	}
	if opts.QuickResolver != nil {
		cmd.AddCommand(NewQuickCommand(opts.QuickResolver, opts.Streams))
	}
	if opts.ConfigResolver != nil {
		cmd.AddCommand(NewConfigCommand(opts.ConfigResolver, opts.Streams))
	}
	return cmd
}

func newVersionCommand(version string, resolveRegion RegionResolver, streams Streams) *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the tick version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !verbose {
				_, err := fmt.Fprintln(streams.Out, version)
				return err
			}
			region := "ticktick"
			if resolveRegion != nil {
				var err error
				region, err = resolveRegion()
				if err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintf(streams.Out, "version: %s\n", version); err != nil {
				return err
			}
			_, err := fmt.Fprintf(streams.Out, "region: %s\n", region)
			return err
		},
	}
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Print runtime details")
	return cmd
}
