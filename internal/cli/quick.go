package cli

import (
	"errors"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/jeely/ticktick-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewQuickCommand(resolveQuickApp QuickResolver, streams Streams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quick",
		Short: "Fast task capture",
	}
	add := &cobra.Command{
		Use:   "add <text>",
		Short: "Create a task from quick-add syntax",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if resolveQuickApp == nil {
				return errors.New("quick command is unavailable")
			}
			quickApp, err := resolveQuickApp()
			if err != nil {
				return err
			}
			task, err := quickApp.Add(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(streams.Out, task)
		},
	}
	cmd.AddCommand(add)
	return cmd
}

type QuickResolver func() (*app.QuickAddApp, error)
