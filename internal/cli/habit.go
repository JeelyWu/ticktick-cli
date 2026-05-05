package cli

import (
	"errors"
	"fmt"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/jeely/ticktick-cli/internal/domain"
	"github.com/jeely/ticktick-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewHabitCommand(resolveHabitApp HabitResolver, resolveConfigApp ConfigResolver, streams Streams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "habit",
		Short: "Read and write TickTick habits",
	}

	resolve := func() (*app.HabitApp, error) {
		if resolveHabitApp == nil {
			return nil, errors.New("habit command is unavailable")
		}
		return resolveHabitApp()
	}

	var lsJSON bool
	ls := &cobra.Command{
		Use:   "ls",
		Short: "List habits",
		RunE: func(cmd *cobra.Command, args []string) error {
			habitApp, err := resolve()
			if err != nil {
				return err
			}
			habits, err := habitApp.List(cmd.Context())
			if err != nil {
				return err
			}
			format, err := resolveOutputFormat(cmd, resolveConfigApp, lsJSON, "")
			if err != nil {
				return err
			}
			if format == "json" {
				return output.PrintJSON(streams.Out, habits)
			}
			return output.PrintHabitsTable(streams.Out, habits)
		},
	}
	ls.Flags().BoolVar(&lsJSON, "json", false, "Print JSON")

	var getJSON bool
	get := &cobra.Command{
		Use:   "get <habit>",
		Short: "Show one habit by exact ID or name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			habitApp, err := resolve()
			if err != nil {
				return err
			}
			habit, err := habitApp.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			format, err := resolveOutputFormat(cmd, resolveConfigApp, getJSON, "")
			if err != nil {
				return err
			}
			if format == "json" {
				return output.PrintJSON(streams.Out, habit)
			}
			return output.PrintHabitsTable(streams.Out, []domain.Habit{habit})
		},
	}
	get.Flags().BoolVar(&getJSON, "json", false, "Print JSON")

	var addInput domain.CreateHabitInput
	add := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a habit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			habitApp, err := resolve()
			if err != nil {
				return err
			}
			addInput.Name = args[0]
			habit, err := habitApp.Add(cmd.Context(), addInput)
			if err != nil {
				return err
			}
			return output.PrintJSON(streams.Out, habit)
		},
	}
	add.Flags().IntVar(&addInput.Goal, "goal", 0, "Daily goal")
	add.Flags().StringVar(&addInput.Color, "color", "", "Color hex")
	add.Flags().StringVar(&addInput.Icon, "icon", "", "Icon")
	add.Flags().StringVar(&addInput.RepeatRule, "repeat", "", "Repeat rule")
	add.Flags().IntSliceVar(&addInput.TargetDays, "target-days", nil, "Target days")
	add.Flags().StringVar(&addInput.Unit, "unit", "", "Unit")
	add.Flags().IntVar(&addInput.Step, "step", 0, "Step")

	var updateInput domain.UpdateHabitInput
	update := &cobra.Command{
		Use:   "update <habit>",
		Short: "Update a habit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			habitApp, err := resolve()
			if err != nil {
				return err
			}
			updateInput.Reference = args[0]
			habit, err := habitApp.Update(cmd.Context(), updateInput)
			if err != nil {
				return err
			}
			return output.PrintJSON(streams.Out, habit)
		},
	}
	update.Flags().StringVar(&updateInput.Name, "name", "", "New name")
	update.Flags().IntVar(&updateInput.Goal, "goal", 0, "New goal")
	update.Flags().StringVar(&updateInput.Color, "color", "", "New color")
	update.Flags().StringVar(&updateInput.Icon, "icon", "", "New icon")
	update.Flags().StringVar(&updateInput.RepeatRule, "repeat", "", "New repeat rule")
	update.Flags().IntSliceVar(&updateInput.TargetDays, "target-days", nil, "New target days")
	update.Flags().StringVar(&updateInput.Unit, "unit", "", "New unit")
	update.Flags().IntVar(&updateInput.Step, "step", 0, "New step")

	archive := &cobra.Command{
		Use:   "archive <habit>",
		Short: "Archive or unarchive a habit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			habitApp, err := resolve()
			if err != nil {
				return err
			}
			habit, err := habitApp.Archive(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(streams.Out, "%s\n", habit.Status.String())
			return err
		},
	}

	var yes bool
	rm := &cobra.Command{
		Use:   "rm <habit>",
		Short: "Delete a habit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			habitApp, err := resolve()
			if err != nil {
				return err
			}
			if !yes {
				ok, err := Confirm(streams, "Delete habit "+args[0]+"?")
				if err != nil {
					return err
				}
				if !ok {
					_, err = fmt.Fprintln(streams.Out, "Cancelled")
					return err
				}
			}
			if err := habitApp.Remove(cmd.Context(), args[0]); err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, "Deleted")
			return err
		},
	}
	rm.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")

	var checkinValue int
	checkin := &cobra.Command{
		Use:   "checkin <habit>",
		Short: "Check in a habit today",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			habitApp, err := resolve()
			if err != nil {
				return err
			}
			if err := habitApp.Checkin(cmd.Context(), args[0], checkinValue); err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, "Checked in")
			return err
		},
	}
	checkin.Flags().IntVar(&checkinValue, "value", 0, "Check-in value")

	var logJSON bool
	logCmd := &cobra.Command{
		Use:   "log <habit>",
		Short: "Show check-in history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			habitApp, err := resolve()
			if err != nil {
				return err
			}
			checkins, err := habitApp.Log(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			format, err := resolveOutputFormat(cmd, resolveConfigApp, logJSON, "")
			if err != nil {
				return err
			}
			if format == "json" {
				return output.PrintJSON(streams.Out, checkins)
			}
			return output.PrintCheckinsTable(streams.Out, checkins)
		},
	}
	logCmd.Flags().BoolVar(&logJSON, "json", false, "Print JSON")

	cmd.AddCommand(ls, get, add, update, archive, rm, checkin, logCmd)
	return cmd
}
