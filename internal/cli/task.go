package cli

import (
	"errors"
	"fmt"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/jeely/ticktick-cli/internal/domain"
	"github.com/jeely/ticktick-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewTaskCommand(resolveTaskApp TaskResolver, streams Streams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Read and write TickTick tasks",
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
	var today bool
	var overdue bool
	ls := &cobra.Command{
		Use:   "ls",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			if today && overdue {
				return errors.New("--today and --overdue cannot be used together")
			}
			if status == "completed" && today {
				return errors.New("--today requires open tasks")
			}
			if status == "completed" && overdue {
				return errors.New("--overdue requires open tasks")
			}
			taskApp, err := resolve()
			if err != nil {
				return err
			}
			input := app.ListTasksInput{
				Project: project,
				From:    from,
				To:      to,
				Today:   today,
				Overdue: overdue,
			}
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
	ls.Flags().BoolVar(&today, "today", false, "Only tasks due today or overdue")
	ls.Flags().BoolVar(&overdue, "overdue", false, "Only overdue tasks")
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

	var addInput domain.CreateTaskInput
	var addPriority int
	add := &cobra.Command{
		Use:   "add <title>",
		Short: "Create a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskApp, err := resolve()
			if err != nil {
				return err
			}
			addInput.Title = args[0]
			addInput.Priority = domain.Priority(addPriority)
			task, err := taskApp.Add(cmd.Context(), addInput)
			if err != nil {
				return err
			}
			return output.PrintJSON(streams.Out, task)
		},
	}
	add.Flags().StringVar(&addInput.ProjectRef, "project", "", "Project ID or exact name")
	add.Flags().StringVar(&addInput.Content, "content", "", "Task content")
	add.Flags().StringVar(&addInput.Description, "desc", "", "Task description")
	add.Flags().StringVar(&addInput.StartRaw, "start", "", "Start date (YYYY-MM-DD or RFC3339)")
	add.Flags().StringVar(&addInput.DueRaw, "due", "", "Due date (YYYY-MM-DD or RFC3339)")
	add.Flags().BoolVar(&addInput.AllDay, "all-day", false, "Mark the task as all-day")
	add.Flags().IntVar(&addPriority, "priority", 0, "Priority value: 0, 1, 3, or 5")

	var updateInput domain.UpdateTaskInput
	var updatePriority int
	update := &cobra.Command{
		Use:   "update <task>",
		Short: "Update an open task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskApp, err := resolve()
			if err != nil {
				return err
			}
			updateInput.Reference = args[0]
			if cmd.Flags().Changed("all-day") {
				value, err := cmd.Flags().GetBool("all-day")
				if err != nil {
					return err
				}
				updateInput.AllDay = &value
			}
			if cmd.Flags().Changed("priority") {
				priority := domain.Priority(updatePriority)
				updateInput.Priority = &priority
			}
			task, err := taskApp.Update(cmd.Context(), updateInput)
			if err != nil {
				return err
			}
			return output.PrintJSON(streams.Out, task)
		},
	}
	update.Flags().StringVar(&updateInput.ProjectRef, "project", "", "Project ID or exact name")
	update.Flags().StringVar(&updateInput.Title, "title", "", "New task title")
	update.Flags().StringVar(&updateInput.Content, "content", "", "New task content")
	update.Flags().StringVar(&updateInput.Description, "desc", "", "New task description")
	update.Flags().StringVar(&updateInput.StartRaw, "start", "", "Start date (YYYY-MM-DD or RFC3339)")
	update.Flags().StringVar(&updateInput.DueRaw, "due", "", "Due date (YYYY-MM-DD or RFC3339)")
	update.Flags().Bool("all-day", false, "Mark the task as all-day")
	update.Flags().IntVar(&updatePriority, "priority", 0, "Priority value: 0, 1, 3, or 5")

	var doneProject string
	done := &cobra.Command{
		Use:   "done <task>",
		Short: "Mark an open task as completed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskApp, err := resolve()
			if err != nil {
				return err
			}
			if err := taskApp.Done(cmd.Context(), args[0], doneProject); err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, "Completed")
			return err
		},
	}
	done.Flags().StringVar(&doneProject, "project", "", "Project ID or exact name")

	var removeProject string
	var yes bool
	rm := &cobra.Command{
		Use:   "rm <task>",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskApp, err := resolve()
			if err != nil {
				return err
			}
			if !yes {
				ok, err := Confirm(streams, "Delete task "+args[0]+"?")
				if err != nil {
					return err
				}
				if !ok {
					_, err = fmt.Fprintln(streams.Out, "Cancelled")
					return err
				}
			}
			if err := taskApp.Remove(cmd.Context(), args[0], removeProject); err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, "Deleted")
			return err
		},
	}
	rm.Flags().StringVar(&removeProject, "project", "", "Project ID or exact name")
	rm.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")

	var fromProject string
	var toProject string
	move := &cobra.Command{
		Use:   "move <task>",
		Short: "Move a task to another project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskApp, err := resolve()
			if err != nil {
				return err
			}
			if err := taskApp.Move(cmd.Context(), domain.MoveTaskInput{
				Reference:      args[0],
				FromProjectRef: fromProject,
				ToProjectRef:   toProject,
			}); err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, "Moved")
			return err
		},
	}
	move.Flags().StringVar(&fromProject, "project", "", "Current project ID or exact name")
	move.Flags().StringVar(&toProject, "to", "", "Destination project ID or exact name")
	_ = move.MarkFlagRequired("to")

	cmd.AddCommand(ls, get, add, update, done, rm, move)
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
