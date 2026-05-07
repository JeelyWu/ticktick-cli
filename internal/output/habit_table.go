package output

import (
	"fmt"
	"io"
	"strconv"
	"text/tabwriter"

	"github.com/jeely/ticktick-cli/internal/domain"
)

func PrintHabitsTable(w io.Writer, habits []domain.Habit) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tNAME\tGOAL\tSTREAK\tSTATUS")
	for _, habit := range habits {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%g\t%d\t%s\n",
			habit.ID, habit.Name, habit.Goal, habit.CurrentStreak, habit.Status.String())
	}
	return tw.Flush()
}

func PrintCheckinsTable(w io.Writer, checkins []domain.HabitCheckin) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "DATE\tVALUE\tGOAL")
	for _, c := range checkins {
		date := formatCheckinStamp(c.Stamp)
		_, _ = fmt.Fprintf(tw, "%s\t%g\t%g\n", date, c.Value, c.Goal)
	}
	return tw.Flush()
}

func formatCheckinStamp(stamp int) string {
	s := strconv.Itoa(stamp)
	if len(s) != 8 {
		return s
	}
	return s[:4] + "-" + s[4:6] + "-" + s[6:]
}
