package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/jeely/ticktick-cli/internal/config"
)

func TestConfigListPrintsFullConfig(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	cmd := NewConfigCommand(func() (*app.ConfigApp, error) {
		store := config.NewStore(t.TempDir() + "/config.yaml")
		configApp := &app.ConfigApp{Store: store}
		if err := configApp.Set(context.Background(), "output.default", "json"); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
		if err := configApp.Set(context.Background(), "task.default_project", "Work"); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
		return configApp, nil
	}, streams)
	cmd.SetIn(streams.In)
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.ErrOut)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "default_project: Work") {
		t.Fatalf("stdout = %q, want default_project", stdout.String())
	}
	if !strings.Contains(stdout.String(), "default: json") {
		t.Fatalf("stdout = %q, want output.default", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}
