package cli

import (
	"errors"
	"fmt"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/jeely/ticktick-cli/internal/domain"
	"github.com/jeely/ticktick-cli/internal/output"
	"github.com/spf13/cobra"
)

type FocusResolver func() (*app.FocusApp, error)

func NewFocusCommand(resolveFocusApp FocusResolver, resolveConfigApp ConfigResolver, streams Streams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focus",
		Short: "Manage TickTick focus sessions",
	}

	resolve := func() (*app.FocusApp, error) {
		if resolveFocusApp == nil {
			return nil, errors.New("focus command is unavailable")
		}
		return resolveFocusApp()
	}

	var from string
	var to string
	var project string
	var jsonOut bool
	var outputFormat string
	ls := &cobra.Command{
		Use:   "ls",
		Short: "List focus sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			focusApp, err := resolve()
			if err != nil {
				return err
			}
			input := app.ListFocusInput{
				From:    from,
				To:      to,
				Project: project,
			}
			focuses, names, err := focusApp.List(cmd.Context(), input)
			if err != nil {
				return err
			}
			format, err := resolveOutputFormat(cmd, resolveConfigApp, jsonOut, "output")
			if err != nil {
				return err
			}
			if format == "json" {
				return output.PrintJSON(streams.Out, focuses)
			}
			return output.PrintFocusTable(streams.Out, focuses, names)
		},
	}
	ls.Flags().StringVar(&from, "from", "", "Start date (YYYY-MM-DD or RFC3339)")
	ls.Flags().StringVar(&to, "to", "", "End date (YYYY-MM-DD or RFC3339)")
	ls.Flags().StringVar(&project, "project", "", "Project ID or exact name")
	ls.Flags().StringVar(&outputFormat, "output", "table", "Output format: table or json")
	ls.Flags().BoolVar(&jsonOut, "json", false, "Print JSON")

	var getJSON bool
	get := &cobra.Command{
		Use:   "get <focus-id>",
		Short: "Show one focus session by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			focusApp, err := resolve()
			if err != nil {
				return err
			}
			focus, err := focusApp.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			format, err := resolveOutputFormat(cmd, resolveConfigApp, getJSON, "")
			if err != nil {
				return err
			}
			if format == "json" {
				return output.PrintJSON(streams.Out, focus)
			}
			names := map[string]string{focus.ProjectID: ""}
			return output.PrintFocusTable(streams.Out, []domain.Focus{focus}, names)
		},
	}
	get.Flags().BoolVar(&getJSON, "json", false, "Print JSON")

	var startContent string
	var startProject string
	var startMode int
	var startTaskID string
	var startTime string
	var startJSON bool
	start := &cobra.Command{
		Use:   "start <title>",
		Short: "Start a focus session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			focusApp, err := resolve()
			if err != nil {
				return err
			}

			input := app.StartFocusAppInput{
				Title:   args[0],
				Content: startContent,
				Mode:    domain.FocusMode(startMode),
				TaskID:  startTaskID,
				StartRaw: startTime,
			}

			if startProject == "" && resolveConfigApp != nil {
				configApp, err := resolveConfigApp()
				if err != nil {
					return err
				}
				defaultProject, err := configApp.Get(cmd.Context(), "default_project")
				if err != nil {
					return err
				}
				startProject = defaultProject
			}

			if startProject == "" {
				if !IsTerminal(streams) {
					return errors.New("no project specified: use --project or set default_project")
				}
				projects, err := focusApp.ListProjects(cmd.Context())
				if err != nil {
					return err
				}
				if len(projects) == 0 {
					return errors.New("no projects available")
				}
				project, err := SelectProject(streams, projects)
				if err != nil {
					return err
				}
				startProject = project.Name
			}

			input.ProjectRef = startProject
			focus, err := focusApp.Start(cmd.Context(), input)
			if err != nil {
				return err
			}
			if startJSON {
				return output.PrintJSON(streams.Out, focus)
			}
			_, err = fmt.Fprintf(streams.Out, "Started focus: %s (%s)\n", focus.ID, focus.Title)
			return err
		},
	}
	start.Flags().StringVar(&startContent, "content", "", "Focus session description")
	start.Flags().StringVar(&startProject, "project", "", "Project ID or exact name")
	start.Flags().IntVar(&startMode, "mode", 1, "Focus mode: 1=timer, 2=pomodoro")
	start.Flags().StringVar(&startTaskID, "task", "", "Associated task ID")
	start.Flags().StringVar(&startTime, "start", "", "Start time (YYYY-MM-DD or RFC3339)")
	start.Flags().BoolVar(&startJSON, "json", false, "Print JSON")

	stop := &cobra.Command{
		Use:   "stop <focus-id>",
		Short: "Stop a focus session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			focusApp, err := resolve()
			if err != nil {
				return err
			}
			if err := focusApp.Stop(cmd.Context(), args[0]); err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, "Stopped")
			return err
		},
	}

	cmd.AddCommand(ls, get, start, stop)
	return cmd
}
