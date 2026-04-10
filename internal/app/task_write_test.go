package app

import (
	"testing"

	"github.com/jeely/ticktick-cli/internal/domain"
)

func TestResolveTaskReferenceByTitleAmbiguous(t *testing.T) {
	tasks := []domain.Task{
		{ID: "a1", Title: "Spec"},
		{ID: "b2", Title: "Spec"},
	}

	if _, err := resolveTaskReference("Spec", tasks); err == nil {
		t.Fatal("resolveTaskReference() error = nil, want ambiguity error")
	}
}
