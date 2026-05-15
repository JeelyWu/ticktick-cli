package cli

import (
	"errors"

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
	var focusType int
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
				From: from,
				To:   to,
				Type: focusType,
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
	ls.Flags().IntVar(&focusType, "type", 1, "Focus type: 0=pomodoro, 1=timer")
	ls.Flags().StringVar(&outputFormat, "output", "table", "Output format: table or json")
	ls.Flags().BoolVar(&jsonOut, "json", false, "Print JSON")

	var getJSON bool
	var getType int
	get := &cobra.Command{
		Use:   "get <focus-id>",
		Short: "Show one focus session by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			focusApp, err := resolve()
			if err != nil {
				return err
			}
			focus, err := focusApp.Get(cmd.Context(), args[0], getType)
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
	get.Flags().IntVar(&getType, "type", 1, "Focus type: 0=pomodoro, 1=timer")
	get.Flags().BoolVar(&getJSON, "json", false, "Print JSON")

	cmd.AddCommand(ls, get)
	return cmd
}
