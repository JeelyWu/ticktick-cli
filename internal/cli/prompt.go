package cli

import (
	"bufio"
	"fmt"
	"strings"
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
