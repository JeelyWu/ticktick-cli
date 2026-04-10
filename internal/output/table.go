package output

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/jeely/ticktick-cli/internal/domain"
)

func PrintProjectsTable(w io.Writer, projects []domain.Project) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tNAME\tCLOSED\tKIND")
	for _, project := range projects {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%t\t%s\n", project.ID, project.Name, project.Closed, project.Kind)
	}
	return tw.Flush()
}

func PrintTasksTable(w io.Writer, tasks []domain.Task, projectNames map[string]string) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tTITLE\tPROJECT\tDUE\tPRIORITY\tSTATUS")
	for _, task := range tasks {
		_, _ = fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\t%s\t%s\n",
			task.ID,
			task.Title,
			projectNames[task.ProjectID],
			FormatTime(task.DueDate),
			task.Priority.String(),
			task.Status.String(),
		)
	}
	return tw.Flush()
}
