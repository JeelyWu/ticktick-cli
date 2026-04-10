package cli

import (
	"errors"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/jeely/ticktick-cli/internal/domain"
	"github.com/jeely/ticktick-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewTaskCommand(resolveTaskApp TaskResolver, streams Streams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Read TickTick tasks",
	}

	resolve := func() (*app.TaskApp, error) {
		if resolveTaskApp == nil {
			return nil, errors.New("task command is unavailable")
		}
		return resolveTaskApp()
	}

	var project string
	var tag string
	var jsonOut bool
	var outputFormat string
	var status string
	var priority int
	var from string
	var to string
	ls := &cobra.Command{
		Use:   "ls",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			taskApp, err := resolve()
			if err != nil {
				return err
			}
			input := app.ListTasksInput{Project: project, From: from, To: to}
			if tag != "" {
				input.Tags = []string{tag}
			}
			if status == "" || status == "open" {
				input.Statuses = []domain.TaskStatus{domain.StatusOpen}
			}
			if status == "completed" {
				input.Statuses = []domain.TaskStatus{domain.StatusCompleted}
			}
			if priority != 0 {
				input.Priorities = []domain.Priority{domain.Priority(priority)}
			}
			tasks, names, err := taskApp.List(cmd.Context(), input)
			if err != nil {
				return err
			}
			if jsonOut || outputFormat == "json" {
				return output.PrintJSON(streams.Out, tasks)
			}
			return output.PrintTasksTable(streams.Out, tasks, names)
		},
	}
	ls.Flags().StringVar(&project, "project", "", "Project ID or exact name")
	ls.Flags().StringVar(&tag, "tag", "", "Tag filter")
	ls.Flags().StringVar(&status, "status", "open", "Task status: open or completed")
	ls.Flags().IntVar(&priority, "priority", 0, "Priority value: 0, 1, 3, or 5")
	ls.Flags().StringVar(&from, "from", "", "Filter start date (YYYY-MM-DD or RFC3339)")
	ls.Flags().StringVar(&to, "to", "", "Filter end date (YYYY-MM-DD or RFC3339)")
	ls.Flags().StringVar(&outputFormat, "output", "table", "Output format: table or json")
	ls.Flags().BoolVar(&jsonOut, "json", false, "Print JSON")

	var getProject string
	var getJSON bool
	get := &cobra.Command{
		Use:   "get <task>",
		Short: "Show one open task by exact ID or title",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskApp, err := resolve()
			if err != nil {
				return err
			}
			task, names, err := taskApp.Get(cmd.Context(), args[0], getProject)
			if err != nil {
				return err
			}
			if getJSON {
				return output.PrintJSON(streams.Out, task)
			}
			return output.PrintTasksTable(streams.Out, []domain.Task{task}, names)
		},
	}
	get.Flags().StringVar(&getProject, "project", "", "Project ID or exact name")
	get.Flags().BoolVar(&getJSON, "json", false, "Print JSON")

	cmd.AddCommand(ls, get)
	return cmd
}

func NewTodayCommand(resolveTaskApp TaskResolver, streams Streams) *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "today",
		Short: "Show tasks due today or overdue",
		RunE: func(cmd *cobra.Command, args []string) error {
			if resolveTaskApp == nil {
				return errors.New("today command is unavailable")
			}
			taskApp, err := resolveTaskApp()
			if err != nil {
				return err
			}
			tasks, names, err := taskApp.Today(cmd.Context())
			if err != nil {
				return err
			}
			if jsonOut {
				return output.PrintJSON(streams.Out, tasks)
			}
			return output.PrintTasksTable(streams.Out, tasks, names)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON")
	return cmd
}

func NewInboxCommand(resolveTaskApp TaskResolver, streams Streams) *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "inbox",
		Short: "Show tasks from the configured inbox project",
		RunE: func(cmd *cobra.Command, args []string) error {
			if resolveTaskApp == nil {
				return errors.New("inbox command is unavailable")
			}
			taskApp, err := resolveTaskApp()
			if err != nil {
				return err
			}
			tasks, names, err := taskApp.Inbox(cmd.Context())
			if err != nil {
				return err
			}
			if jsonOut {
				return output.PrintJSON(streams.Out, tasks)
			}
			return output.PrintTasksTable(streams.Out, tasks, names)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON")
	return cmd
}
