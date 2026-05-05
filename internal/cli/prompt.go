package cli

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/jeely/ticktick-cli/internal/domain"
	"golang.org/x/term"
)

func Confirm(streams Streams, prompt string) (bool, error) {
	_, err := fmt.Fprintf(streams.Out, "%s [y/N]: ", prompt)
	if err != nil {
		return false, err
	}
	answer, err := bufio.NewReader(streams.In).ReadString('\n')
	if err != nil {
		return false, err
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes", nil
}

func IsTerminal(streams Streams) bool {
	f, ok := streams.In.(interface{ Fd() uintptr })
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func SelectProject(streams Streams, projects []domain.Project) (domain.Project, error) {
	for i, p := range projects {
		_, _ = fmt.Fprintf(streams.Out, "  %d. %s\n", i+1, p.Name)
	}
	reader := bufio.NewReader(streams.In)
	for {
		_, err := fmt.Fprint(streams.Out, "Select project: ")
		if err != nil {
			return domain.Project{}, err
		}
		input, err := reader.ReadString('\n')
		if err != nil {
			return domain.Project{}, err
		}
		input = strings.TrimSpace(input)
		idx, err := strconv.Atoi(input)
		if err != nil || idx < 1 || idx > len(projects) {
			_, _ = fmt.Fprintln(streams.Out, "Invalid selection")
			continue
		}
		return projects[idx-1], nil
	}
}
