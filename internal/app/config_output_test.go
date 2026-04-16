package app

import (
	"context"
	"strings"
	"testing"

	"github.com/jeely/ticktick-cli/internal/config"
)

func TestConfigAppSetRejectsUnsupportedOutputDefault(t *testing.T) {
	store := config.NewStore(t.TempDir() + "/config.yaml")
	app := ConfigApp{Store: store}

	err := app.Set(context.Background(), "output.default", "yaml")
	if err == nil {
		t.Fatal("Set() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "output.default") {
		t.Fatalf("error = %q, want output.default validation", err)
	}
}
