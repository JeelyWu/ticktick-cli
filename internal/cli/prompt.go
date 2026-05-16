package cli

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/jeelywu/ticktick-cli/internal/domain"
	"golang.org/x/term"
)

func getReader(streams Streams) *bufio.Reader {
	if r, ok := streams.In.(*bufio.Reader); ok {
		return r
	}
	return bufio.NewReader(streams.In)
}

func Confirm(streams Streams, prompt string) (bool, error) {
	_, err := fmt.Fprintf(streams.Out, "%s [y/N]: ", prompt)
	if err != nil {
		return false, err
	}
	answer, err := getReader(streams).ReadString('\n')
	if err != nil {
		return false, err
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes", nil
}

var isTerminal = func(streams Streams) bool {
	f, ok := streams.In.(interface{ Fd() uintptr })
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func IsTerminal(streams Streams) bool {
	return isTerminal(streams)
}

func SelectProject(streams Streams, projects []domain.Project) (domain.Project, error) {
	for i, p := range projects {
		_, _ = fmt.Fprintf(streams.Out, "  %d. %s\n", i+1, p.Name)
	}
	reader := getReader(streams)
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

func SelectRegion(streams Streams, defaultRegion string) (string, error) {
	_, _ = fmt.Fprintln(streams.Out, "  1. ticktick (International version)")
	_, _ = fmt.Fprintln(streams.Out, "  2. dida365 (China version)")

	var defaultIdx int
	switch defaultRegion {
	case "ticktick":
		defaultIdx = 1
	case "dida365":
		defaultIdx = 2
	}

	reader := getReader(streams)
	for {
		if defaultIdx > 0 {
			_, err := fmt.Fprintf(streams.Out, "Select [%d]: ", defaultIdx)
			if err != nil {
				return "", err
			}
		} else {
			_, err := fmt.Fprint(streams.Out, "Select: ")
			if err != nil {
				return "", err
			}
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		input = strings.TrimSpace(input)

		if input == "" && defaultIdx > 0 {
			switch defaultIdx {
			case 1:
				return "ticktick", nil
			case 2:
				return "dida365", nil
			}
		}

		idx, err := strconv.Atoi(input)
		if err != nil || idx < 1 || idx > 2 {
			_, _ = fmt.Fprintln(streams.Out, "Invalid selection")
			continue
		}
		switch idx {
		case 1:
			return "ticktick", nil
		case 2:
			return "dida365", nil
		}
	}
}

func Prompt(streams Streams, label, defaultValue string) (string, error) {
	if defaultValue != "" {
		_, err := fmt.Fprintf(streams.Out, "%s [%s]: ", label, defaultValue)
		if err != nil {
			return "", err
		}
	} else {
		_, err := fmt.Fprintf(streams.Out, "%s: ", label)
		if err != nil {
			return "", err
		}
	}

	input, err := getReader(streams).ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}
	return input, nil
}

func PromptSecret(streams Streams, label string, hasExisting bool) (string, error) {
	if hasExisting {
		_, err := fmt.Fprintf(streams.Out, "%s (press Enter to keep existing): ", label)
		if err != nil {
			return "", err
		}
	} else {
		_, err := fmt.Fprintf(streams.Out, "%s: ", label)
		if err != nil {
			return "", err
		}
	}

	input, err := getReader(streams).ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)
	return input, nil
}
