package domain

import (
	"errors"
	"fmt"
)

var ErrNotAuthenticated = errors.New("tick auth login is required")

type ReferenceError struct {
	Kind      string
	Reference string
	Matches   []string
}

func (e *ReferenceError) Error() string {
	if len(e.Matches) == 0 {
		return fmt.Sprintf("%s %q not found", e.Kind, e.Reference)
	}
	return fmt.Sprintf("%s %q is ambiguous: %v", e.Kind, e.Reference, e.Matches)
}
