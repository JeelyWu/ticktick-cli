package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jeely/ticktick-cli/internal/auth"
	"github.com/jeely/ticktick-cli/internal/config"
)

type fakeAuthService struct {
	status   auth.Status
	loginErr error
}

func (f fakeAuthService) Login(context.Context, auth.LoginInput) (auth.Token, error) {
	if f.loginErr != nil {
		return auth.Token{}, f.loginErr
	}
	return auth.Token{AccessToken: "access-1"}, nil
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

func TestAuthAppLoginDoesNotPersistConfigOnFailure(t *testing.T) {
	store := config.NewStore(t.TempDir() + "/config.yaml")
	app := AuthApp{
		ConfigStore: store,
		Service: fakeAuthService{
			loginErr: errors.New("oauth exchange failed"),
		},
	}

	err := app.Login(context.Background(), LoginInput{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  "http://localhost:14573/callback",
	})
	if err == nil {
		t.Fatal("Login() error = nil, want non-nil")
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.OAuth.ClientID != "" {
		t.Fatalf("OAuth.ClientID = %q, want empty", cfg.OAuth.ClientID)
	}
	if cfg.OAuth.RedirectURL != "" {
		t.Fatalf("OAuth.RedirectURL = %q, want empty", cfg.OAuth.RedirectURL)
	}
}

func TestAuthAppLoginPersistsConfigOnSuccess(t *testing.T) {
	store := config.NewStore(t.TempDir() + "/config.yaml")
	app := AuthApp{
		ConfigStore: store,
		Service:     fakeAuthService{},
	}

	if err := app.Login(context.Background(), LoginInput{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  "http://localhost:14573/callback",
	}); err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.OAuth.ClientID != "client-1" {
		t.Fatalf("OAuth.ClientID = %q, want client-1", cfg.OAuth.ClientID)
	}
	if cfg.OAuth.RedirectURL != "http://localhost:14573/callback" {
		t.Fatalf("OAuth.RedirectURL = %q, want callback", cfg.OAuth.RedirectURL)
	}
}

func TestAuthAppLoginSucceedsWithoutConfigStoreWhenInputsAreExplicit(t *testing.T) {
	app := AuthApp{
		Service: fakeAuthService{},
	}

	if err := app.Login(context.Background(), LoginInput{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  "http://localhost:14573/callback",
	}); err != nil {
		t.Fatalf("Login() error = %v", err)
	}
}

func TestAuthAppLoginFailsWithoutConfigStoreWhenDefaultsAreRequired(t *testing.T) {
	app := AuthApp{
		Service: fakeAuthService{},
	}

	err := app.Login(context.Background(), LoginInput{
		ClientSecret: "secret-1",
	})
	if err == nil {
		t.Fatal("Login() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "client-id") {
		t.Fatalf("error = %q, want client-id message", err)
	}
}
