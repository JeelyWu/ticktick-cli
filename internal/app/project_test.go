package app

import (
	"testing"

	"github.com/jeely/ticktick-cli/internal/domain"
)

func TestResolveProjectByIDAndName(t *testing.T) {
	projects := []domain.Project{
		{ID: "p1", Name: "Inbox"},
		{ID: "p2", Name: "Work"},
	}

	project, err := ResolveProject("Work", projects)
	if err != nil {
		t.Fatalf("ResolveProject() error = %v", err)
	}
	if project.ID != "p2" {
		t.Fatalf("project.ID = %q, want p2", project.ID)
	}

	project, err = ResolveProject("p1", projects)
	if err != nil {
		t.Fatalf("ResolveProject() by id error = %v", err)
	}
	if project.Name != "Inbox" {
		t.Fatalf("project.Name = %q, want Inbox", project.Name)
	}
}

func TestResolveProjectAmbiguous(t *testing.T) {
	projects := []domain.Project{
		{ID: "p1", Name: "Inbox"},
		{ID: "p2", Name: "Inbox"},
	}

	if _, err := ResolveProject("Inbox", projects); err == nil {
		t.Fatal("ResolveProject() error = nil, want ambiguity error")
	}
}
