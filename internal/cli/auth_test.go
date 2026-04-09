package cli

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/jeely/ticktick-cli/internal/auth"
	"github.com/jeely/ticktick-cli/internal/config"
)

type recordingAuthService struct {
	loginInput auth.LoginInput
}

func (r *recordingAuthService) Login(_ context.Context, in auth.LoginInput) (auth.Token, error) {
	r.loginInput = in
	return auth.Token{AccessToken: "access-1"}, nil
}

func (r *recordingAuthService) Status(context.Context) (auth.Status, error) {
	return auth.Status{}, nil
}

func (r *recordingAuthService) Logout(context.Context) error {
	return nil
}

func TestAuthLoginUsesClientSecretFromEnvironment(t *testing.T) {
	t.Setenv("TICK_CLIENT_SECRET", "env-secret")

	streams, stdout, stderr := newTestStreams()
	service := &recordingAuthService{}
	cmd := NewAuthCommand(&app.AuthApp{
		ConfigStore: config.NewStore(t.TempDir() + "/config.yaml"),
		Service:     service,
	}, streams)
	cmd.SetIn(streams.In)
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.ErrOut)
	cmd.SetArgs([]string{
		"login",
		"--client-id", "client-1",
		"--redirect-url", "http://localhost:14573/callback",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if service.loginInput.ClientSecret != "env-secret" {
		t.Fatalf("ClientSecret = %q, want env-secret", service.loginInput.ClientSecret)
	}
	if got := strings.TrimSpace(stdout.String()); got != "Login successful" {
		t.Fatalf("stdout = %q, want Login successful", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestAuthLoginHelpMentionsEnvironmentFallback(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	cmd := NewAuthCommand(&app.AuthApp{
		ConfigStore: config.NewStore(t.TempDir() + "/config.yaml"),
		Service:     &recordingAuthService{},
	}, streams)
	cmd.SetIn(streams.In)
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.ErrOut)
	cmd.SetArgs([]string{"login", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "TICK_CLIENT_SECRET") {
		t.Fatalf("help output = %q, want TICK_CLIENT_SECRET", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestAuthLoginFlagOverridesEnvironment(t *testing.T) {
	t.Setenv("TICK_CLIENT_SECRET", "env-secret")

	streams, _, _ := newTestStreams()
	service := &recordingAuthService{}
	cmd := NewAuthCommand(&app.AuthApp{
		ConfigStore: config.NewStore(t.TempDir() + "/config.yaml"),
		Service:     service,
	}, streams)
	cmd.SetIn(streams.In)
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.ErrOut)
	cmd.SetArgs([]string{
		"login",
		"--client-id", "client-1",
		"--client-secret", "flag-secret",
		"--redirect-url", "http://localhost:14573/callback",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if service.loginInput.ClientSecret != "flag-secret" {
		t.Fatalf("ClientSecret = %q, want flag-secret", service.loginInput.ClientSecret)
	}
}

func TestAuthLoginWithoutEnvOrFlagStillFails(t *testing.T) {
	_ = os.Unsetenv("TICK_CLIENT_SECRET")

	streams, _, _ := newTestStreams()
	cmd := NewAuthCommand(&app.AuthApp{
		ConfigStore: config.NewStore(t.TempDir() + "/config.yaml"),
		Service:     &recordingAuthService{},
	}, streams)
	cmd.SetIn(streams.In)
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.ErrOut)
	cmd.SetArgs([]string{
		"login",
		"--client-id", "client-1",
		"--redirect-url", "http://localhost:14573/callback",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "client-secret") {
		t.Fatalf("error = %q, want client-secret message", err)
	}
}
