package app

import (
	"context"
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
