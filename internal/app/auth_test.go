package app

import (
	"context"
	"testing"

	"github.com/jeely/ticktick-cli/internal/auth"
	"github.com/jeely/ticktick-cli/internal/config"
)

type fakeAuthService struct {
	status auth.Status
}

func (f fakeAuthService) Login(context.Context, auth.LoginInput) (auth.Token, error) {
	return auth.Token{}, nil
}

func (f fakeAuthService) Status(context.Context) (auth.Status, error) {
	return f.status, nil
}

func (f fakeAuthService) Logout(context.Context) error {
	return nil
}

func TestAuthAppStatus(t *testing.T) {
	store := config.NewStore(t.TempDir() + "/config.yaml")
	app := AuthApp{
		ConfigStore: store,
		Service: fakeAuthService{
			status: auth.Status{Authenticated: true},
		},
	}

	status, err := app.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if !status.Authenticated {
		t.Fatal("Authenticated = false, want true")
	}
}
