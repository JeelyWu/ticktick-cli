package cli

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jeelywu/ticktick-cli/internal/app"
	"github.com/jeelywu/ticktick-cli/internal/auth"
	"github.com/jeelywu/ticktick-cli/internal/config"
)

type recordingAuthService struct {
	loginInput  auth.LoginInput
	status      auth.Status
	logoutCalls int
}

func (r *recordingAuthService) Login(_ context.Context, in auth.LoginInput) (auth.Token, error) {
	r.loginInput = in
	return auth.Token{AccessToken: "access-1"}, nil
}

func (r *recordingAuthService) Status(context.Context) (auth.Status, error) {
	return r.status, nil
}

func (r *recordingAuthService) Logout(context.Context) error {
	r.logoutCalls++
	return nil
}

func TestAuthLogin_Interactive(t *testing.T) {
	// Override IsTerminal to always return true for this test
	origIsTerminal := isTerminal
	isTerminal = func(Streams) bool { return true }
	defer func() { isTerminal = origIsTerminal }()

	streams, stdout, stderr := newTestStreams()
	streams.In = bufio.NewReader(bytes.NewBufferString("1\nmy-client-id\nmy-secret\n"))

	service := &recordingAuthService{}
	store := config.NewStore(t.TempDir() + "/config.yaml")
	cmd := NewAuthCommand(func() (*app.AuthApp, error) {
		return &app.AuthApp{
			ConfigStore: store,
			Service:     service,
		}, nil
	}, nil, nil, streams)
	cmd.SetIn(streams.In)
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.ErrOut)
	cmd.SetArgs([]string{"login"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if service.loginInput.ClientID != "my-client-id" {
		t.Fatalf("ClientID = %q, want my-client-id", service.loginInput.ClientID)
	}
	if service.loginInput.ClientSecret != "my-secret" {
		t.Fatalf("ClientSecret = %q, want my-secret", service.loginInput.ClientSecret)
	}
	if !strings.Contains(stdout.String(), "Login successful") {
		t.Fatalf("stdout = %q, want Login successful", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestAuthLogin_UsesConfigDefaults(t *testing.T) {
	// Override IsTerminal to always return true for this test
	origIsTerminal := isTerminal
	isTerminal = func(Streams) bool { return true }
	defer func() { isTerminal = origIsTerminal }()

	streams, stdout, stderr := newTestStreams()
	streams.In = bufio.NewReader(bytes.NewBufferString("\n\n\n"))

	store := config.NewStore(t.TempDir() + "/config.yaml")
	_ = store.Save(config.Config{
		Region:       "dida365",
		ClientID:     "saved-client-id",
		ClientSecret: "saved-secret",
	})

	service := &recordingAuthService{}
	cmd := NewAuthCommand(func() (*app.AuthApp, error) {
		return &app.AuthApp{
			ConfigStore: store,
			Service:     service,
		}, nil
	}, nil, nil, streams)
	cmd.SetIn(streams.In)
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.ErrOut)
	cmd.SetArgs([]string{"login"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if service.loginInput.ClientID != "saved-client-id" {
		t.Fatalf("ClientID = %q, want saved-client-id", service.loginInput.ClientID)
	}
	if service.loginInput.ClientSecret != "saved-secret" {
		t.Fatalf("ClientSecret = %q, want saved-secret", service.loginInput.ClientSecret)
	}
	if !strings.Contains(stdout.String(), "Login successful") {
		t.Fatalf("stdout = %q, want Login successful", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestAuthLogin_NonTerminalFails(t *testing.T) {
	streams, _, _ := newTestStreams()
	// streams.In is a bytes.Buffer, not a terminal
	cmd := NewAuthCommand(func() (*app.AuthApp, error) {
		return &app.AuthApp{
			ConfigStore: config.NewStore(t.TempDir() + "/config.yaml"),
			Service:     &recordingAuthService{},
		}, nil
	}, nil, nil, streams)
	cmd.SetIn(streams.In)
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.ErrOut)
	cmd.SetArgs([]string{"login"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "interactive terminal") {
		t.Fatalf("error = %q, want interactive terminal message", err.Error())
	}
}

func TestAuthHelpDoesNotResolveAuthApp(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	loginResolved := 0
	serviceResolved := 0
	cmd := NewAuthCommand(func() (*app.AuthApp, error) {
		loginResolved++
		return nil, nil
	}, func() (app.AuthService, error) {
		serviceResolved++
		return nil, nil
	}, nil, streams)
	cmd.SetIn(streams.In)
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.ErrOut)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if loginResolved != 0 {
		t.Fatalf("login resolver calls = %d, want 0", loginResolved)
	}
	if serviceResolved != 0 {
		t.Fatalf("service resolver calls = %d, want 0", serviceResolved)
	}
	if !strings.Contains(stdout.String(), "Authenticate with TickTick") {
		t.Fatalf("help output = %q, want auth help", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestAuthStatusUsesServiceResolverWithoutResolvingLoginApp(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	loginResolved := 0
	serviceResolved := 0
	cmd := NewAuthCommand(func() (*app.AuthApp, error) {
		loginResolved++
		return &app.AuthApp{
			ConfigStore: config.NewStore(t.TempDir() + "/config.yaml"),
			Service: &recordingAuthService{
				status: auth.Status{
					Authenticated:    true,
					ExpiryKnown:      true,
					ExpiresAtUnix:    1_776_355_566,
					ExpiresInSeconds: 3600,
				},
			},
		}, nil
	}, func() (app.AuthService, error) {
		serviceResolved++
		return &recordingAuthService{
			status: auth.Status{
				Authenticated:    true,
				ExpiryKnown:      true,
				ExpiresAtUnix:    1_776_355_566,
				ExpiresInSeconds: 3600,
			},
		}, nil
	}, func() (string, error) {
		return "dida365", nil
	}, streams)
	cmd.SetIn(streams.In)
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.ErrOut)
	cmd.SetArgs([]string{"status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if loginResolved != 0 {
		t.Fatalf("login resolver calls = %d, want 0", loginResolved)
	}
	if serviceResolved != 1 {
		t.Fatalf("service resolver calls = %d, want 1", serviceResolved)
	}
	if got := strings.TrimSpace(stdout.String()); got != "authenticated\nregion: dida365\nexpires_at: 2026-04-16T16:06:06Z\nexpires_in: 3600s" {
		t.Fatalf("stdout = %q, want authenticated with expiry details", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestAuthStatusPrintsRegionWhenNotAuthenticated(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	cmd := NewAuthCommand(nil, func() (app.AuthService, error) {
		return &recordingAuthService{
			status: auth.Status{Authenticated: false},
		}, nil
	}, func() (string, error) {
		return "ticktick", nil
	}, streams)
	cmd.SetIn(streams.In)
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.ErrOut)
	cmd.SetArgs([]string{"status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "not authenticated\nregion: ticktick" {
		t.Fatalf("stdout = %q, want not authenticated with region", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestAuthLogoutUsesServiceResolverWithoutResolvingLoginApp(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	loginResolved := 0
	serviceResolved := 0
	service := &recordingAuthService{}
	cmd := NewAuthCommand(func() (*app.AuthApp, error) {
		loginResolved++
		return nil, errors.New("login resolver should not run")
	}, func() (app.AuthService, error) {
		serviceResolved++
		return service, nil
	}, nil, streams)
	cmd.SetIn(streams.In)
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.ErrOut)
	cmd.SetArgs([]string{"logout"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if loginResolved != 0 {
		t.Fatalf("login resolver calls = %d, want 0", loginResolved)
	}
	if serviceResolved != 1 {
		t.Fatalf("service resolver calls = %d, want 1", serviceResolved)
	}
	if service.logoutCalls != 1 {
		t.Fatalf("Logout() calls = %d, want 1", service.logoutCalls)
	}
	if got := strings.TrimSpace(stdout.String()); got != "Logged out" {
		t.Fatalf("stdout = %q, want Logged out", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}
