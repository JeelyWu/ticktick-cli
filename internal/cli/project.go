package cli

import (
	"errors"
	"fmt"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/jeely/ticktick-cli/internal/domain"
	"github.com/jeely/ticktick-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewProjectCommand(resolveProjectApp ProjectResolver, resolveConfigApp ConfigResolver, streams Streams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Read and write TickTick projects",
	}

	resolve := func() (*app.ProjectApp, error) {
		if resolveProjectApp == nil {
			return nil, errors.New("project command is unavailable")
		}
		return resolveProjectApp()
	}

	var lsJSON bool
	ls := &cobra.Command{
		Use:   "ls",
		Short: "List projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectApp, err := resolve()
			if err != nil {
				return err
			}
			projects, err := projectApp.List(cmd.Context())
			if err != nil {
				return err
			}
			format, err := resolveOutputFormat(cmd, resolveConfigApp, lsJSON, "")
			if err != nil {
				return err
			}
			if format == "json" {
				return output.PrintJSON(streams.Out, projects)
			}
			return output.PrintProjectsTable(streams.Out, projects)
		},
	}
	ls.Flags().BoolVar(&lsJSON, "json", false, "Print JSON")

	var getJSON bool
	get := &cobra.Command{
		Use:   "get <project>",
		Short: "Show one project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectApp, err := resolve()
			if err != nil {
				return err
			}
			project, err := projectApp.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			format, err := resolveOutputFormat(cmd, resolveConfigApp, getJSON, "")
			if err != nil {
				return err
			}
			if format == "json" {
				return output.PrintJSON(streams.Out, project)
			}
			return output.PrintProjectsTable(streams.Out, []domain.Project{project})
		},
	}
	get.Flags().BoolVar(&getJSON, "json", false, "Print JSON")

	var addColor string
	var addKind string
	add := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectApp, err := resolve()
			if err != nil {
				return err
			}
			project, err := projectApp.Create(cmd.Context(), args[0], addColor, addKind)
			if err != nil {
				return err
			}
			return output.PrintJSON(streams.Out, project)
		},
	}
	add.Flags().StringVar(&addColor, "color", "", "Project color such as #F18181")
	add.Flags().StringVar(&addKind, "kind", "TASK", "Project kind: TASK or NOTE")

	var updateName string
	var updateColor string
	var updateKind string
	update := &cobra.Command{
		Use:   "update <project>",
		Short: "Update a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectApp, err := resolve()
			if err != nil {
				return err
			}
			project, err := projectApp.Update(cmd.Context(), args[0], updateName, updateColor, updateKind)
			if err != nil {
				return err
			}
			return output.PrintJSON(streams.Out, project)
		},
	}
	update.Flags().StringVar(&updateName, "name", "", "New project name")
	update.Flags().StringVar(&updateColor, "color", "", "New project color")
	update.Flags().StringVar(&updateKind, "kind", "", "New project kind: TASK or NOTE")

	var yes bool
	rm := &cobra.Command{
		Use:   "rm <project>",
		Short: "Delete a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectApp, err := resolve()
			if err != nil {
				return err
			}
			if !yes {
				return fmt.Errorf("project rm requires --yes")
			}
			if err := projectApp.Remove(cmd.Context(), args[0]); err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, "Deleted")
			return err
		},
	}
	rm.Flags().BoolVar(&yes, "yes", false, "Confirm project deletion")

	cmd.AddCommand(ls, get, add, update, rm)
	return cmd
}
