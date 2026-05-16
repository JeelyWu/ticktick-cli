package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jeely/ticktick-cli/internal/auth"
	"github.com/jeely/ticktick-cli/internal/config"
)

type fakeAuthService struct {
	status   auth.Status
	loginErr error
	loginOps *int
}

func (f fakeAuthService) Login(context.Context, auth.LoginInput) (auth.Token, error) {
	if f.loginOps != nil {
		*f.loginOps++
	}
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

type recordingLoginAuthService struct {
	loginInput auth.LoginInput
}

func (r *recordingLoginAuthService) Login(_ context.Context, in auth.LoginInput) (auth.Token, error) {
	r.loginInput = in
	return auth.Token{AccessToken: "access-1"}, nil
}

func (r *recordingLoginAuthService) Status(context.Context) (auth.Status, error) {
	return auth.Status{}, nil
}

func (r *recordingLoginAuthService) Logout(context.Context) error {
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

func TestAuthAppLoginPersistsConfigBeforeServiceLogin(t *testing.T) {
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
		Region:       "dida365",
	})
	if err == nil {
		t.Fatal("Login() error = nil, want non-nil")
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ClientID != "client-1" {
		t.Fatalf("ClientID = %q, want client-1", cfg.ClientID)
	}
	if cfg.ClientSecret != "secret-1" {
		t.Fatalf("ClientSecret = %q, want secret-1", cfg.ClientSecret)
	}
	if cfg.Region != "dida365" {
		t.Fatalf("Region = %q, want dida365", cfg.Region)
	}
	if cfg.RedirectURL != "http://localhost:14573/callback" {
		t.Fatalf("RedirectURL = %q, want callback", cfg.RedirectURL)
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
		Region:       "dida365",
	}); err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ClientID != "client-1" {
		t.Fatalf("ClientID = %q, want client-1", cfg.ClientID)
	}
	if cfg.ClientSecret != "secret-1" {
		t.Fatalf("ClientSecret = %q, want secret-1", cfg.ClientSecret)
	}
	if cfg.Region != "dida365" {
		t.Fatalf("Region = %q, want dida365", cfg.Region)
	}
	if cfg.RedirectURL != "http://localhost:14573/callback" {
		t.Fatalf("RedirectURL = %q, want callback", cfg.RedirectURL)
	}
}

func TestAuthAppLoginBuildsServiceForSelectedRegion(t *testing.T) {
	store := config.NewStore(t.TempDir() + "/config.yaml")
	if err := store.Save(config.Config{
		Region:       "ticktick",
		ClientID:     "old-client-id",
		ClientSecret: "old-secret",
		RedirectURL:  "http://localhost:14573/callback",
	}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	selectedRegion := ""
	service := &recordingLoginAuthService{}
	app := AuthApp{
		ConfigStore: store,
		ServiceForRegion: func(region string) (AuthService, error) {
			selectedRegion = region
			return service, nil
		},
	}

	if err := app.Login(context.Background(), LoginInput{
		Region:       "dida365",
		ClientID:     "new-client-id",
		ClientSecret: "new-secret",
	}); err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if selectedRegion != "dida365" {
		t.Fatalf("service region = %q, want dida365", selectedRegion)
	}
	if service.loginInput.ClientID != "new-client-id" {
		t.Fatalf("ClientID = %q, want new-client-id", service.loginInput.ClientID)
	}
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Region != "dida365" {
		t.Fatalf("stored Region = %q, want dida365", cfg.Region)
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

func TestAuthAppLoginWithExplicitInputsIgnoresMalformedConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("client_id: [\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	store := config.NewStore(path)
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
	if cfg.ClientID != "client-1" {
		t.Fatalf("ClientID = %q, want client-1", cfg.ClientID)
	}
	if cfg.RedirectURL != "http://localhost:14573/callback" {
		t.Fatalf("RedirectURL = %q, want callback", cfg.RedirectURL)
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

func TestAuthAppLoginDefaultsRedirectURL(t *testing.T) {
	store := config.NewStore(t.TempDir() + "/config.yaml")
	app := AuthApp{
		ConfigStore: store,
		Service:     fakeAuthService{},
	}

	if err := app.Login(context.Background(), LoginInput{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
	}); err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.RedirectURL != "http://localhost:8080/callback" {
		t.Fatalf("RedirectURL = %q, want http://localhost:8080/callback", cfg.RedirectURL)
	}
}

func TestAuthAppLoginFillsClientSecretFromConfig(t *testing.T) {
	store := config.NewStore(t.TempDir() + "/config.yaml")
	if err := store.Save(config.Config{
		ClientID:     "client-1",
		ClientSecret: "secret-from-config",
		RedirectURL:  "http://localhost:14573/callback",
		Region:       "ticktick",
	}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	app := AuthApp{
		ConfigStore: store,
		Service:     fakeAuthService{},
	}

	if err := app.Login(context.Background(), LoginInput{
		ClientID: "client-1",
	}); err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ClientSecret != "secret-from-config" {
		t.Fatalf("ClientSecret = %q, want secret-from-config", cfg.ClientSecret)
	}
}

func TestAuthAppLoginFillsRegionFromConfig(t *testing.T) {
	store := config.NewStore(t.TempDir() + "/config.yaml")
	if err := store.Save(config.Config{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  "http://localhost:14573/callback",
		Region:       "dida365",
	}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

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
	if cfg.Region != "dida365" {
		t.Fatalf("Region = %q, want dida365", cfg.Region)
	}
}

func TestAuthAppLoginFailsBeforeServiceCallWhenConfigSaveFails(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "config")
	if err := os.Mkdir(configDir, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	loginOps := 0
	app := AuthApp{
		ConfigStore: config.NewStore(configDir),
		Service: fakeAuthService{
			loginOps: &loginOps,
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
	if loginOps != 0 {
		t.Fatalf("login calls = %d, want 0", loginOps)
	}
}
