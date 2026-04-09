package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/jeely/ticktick-cli/internal/app"
)

func newTestStreams() (Streams, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	return Streams{
		In:     bytes.NewBuffer(nil),
		Out:    stdout,
		ErrOut: stderr,
	}, stdout, stderr
}

func TestRootCommandHelp(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	cmd := NewRootCommand(RootOptions{
		Version: "dev",
		Streams: streams,
	})
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "TickTick CLI") {
		t.Fatalf("help output = %q, want TickTick CLI", stdout.String())
	}
	if strings.Contains(stdout.String(), "completion") {
		t.Fatalf("help output = %q, want no completion command", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestVersionCommandPrintsVersion(t *testing.T) {
	streams, stdout, _ := newTestStreams()
	cmd := NewRootCommand(RootOptions{
		Version: "1.2.3",
		Streams: streams,
	})
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "1.2.3" {
		t.Fatalf("version output = %q, want 1.2.3", got)
	}
}

func TestRootCommandHelpDoesNotResolveAuthApp(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	appResolved := 0
	serviceResolved := 0
	cmd := NewRootCommand(RootOptions{
		Version: "dev",
		Streams: streams,
		LoginAuthResolver: func() (*app.AuthApp, error) {
			appResolved++
			return nil, errors.New("resolver should not run")
		},
		AuthServiceResolver: func() (app.AuthService, error) {
			serviceResolved++
			return nil, errors.New("resolver should not run")
		},
	})
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if appResolved != 0 {
		t.Fatalf("login resolver calls = %d, want 0", appResolved)
	}
	if serviceResolved != 0 {
		t.Fatalf("service resolver calls = %d, want 0", serviceResolved)
	}
	if !strings.Contains(stdout.String(), "auth") {
		t.Fatalf("help output = %q, want auth command listed", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestVersionCommandDoesNotResolveAuthApp(t *testing.T) {
	streams, stdout, stderr := newTestStreams()
	appResolved := 0
	serviceResolved := 0
	cmd := NewRootCommand(RootOptions{
		Version: "1.2.3",
		Streams: streams,
		LoginAuthResolver: func() (*app.AuthApp, error) {
			appResolved++
			return nil, errors.New("resolver should not run")
		},
		AuthServiceResolver: func() (app.AuthService, error) {
			serviceResolved++
			return nil, errors.New("resolver should not run")
		},
	})
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if appResolved != 0 {
		t.Fatalf("login resolver calls = %d, want 0", appResolved)
	}
	if serviceResolved != 0 {
		t.Fatalf("service resolver calls = %d, want 0", serviceResolved)
	}
	if got := strings.TrimSpace(stdout.String()); got != "1.2.3" {
		t.Fatalf("version output = %q, want 1.2.3", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}
