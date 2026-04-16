package app

import (
	"context"
	"strings"
	"testing"

	"github.com/jeely/ticktick-cli/internal/config"
)

func TestConfigAppSetAndGet(t *testing.T) {
	store := config.NewStore(t.TempDir() + "/config.yaml")
	app := ConfigApp{Store: store}

	if err := app.Set(context.Background(), "output.default", "json"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	value, err := app.Get(context.Background(), "output.default")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if value != "json" {
		t.Fatalf("value = %q, want json", value)
	}
}

func TestConfigAppList(t *testing.T) {
	store := config.NewStore(t.TempDir() + "/config.yaml")
	app := ConfigApp{Store: store}

	if err := app.Set(context.Background(), "output.default", "json"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := app.Set(context.Background(), "task.default_project", "Work"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := app.Set(context.Background(), "oauth.client_id", "client-1"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	type configLister interface {
		List(context.Context) (string, error)
	}

	lister, ok := any(app).(configLister)
	if !ok {
		t.Fatal("ConfigApp does not implement List(context.Context)")
	}

	output, err := lister.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	for _, want := range []string{
		"output:",
		"default: json",
		"task:",
		"default_project: Work",
		"inbox_project_id: \"\"",
		"oauth:",
		"client_id: client-1",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("List() output = %q, want substring %q", output, want)
		}
	}
}
