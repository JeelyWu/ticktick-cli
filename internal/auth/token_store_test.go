package auth

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeKeyringBackend struct {
	setErr    error
	getErr    error
	deleteErr error
	values    map[string]string
}

func (b *fakeKeyringBackend) Set(service, user, value string) error {
	if b.setErr != nil {
		return b.setErr
	}
	if b.values == nil {
		b.values = map[string]string{}
	}
	b.values[user] = value
	return nil
}

func (b *fakeKeyringBackend) Get(service, user string) (string, error) {
	if b.getErr != nil {
		return "", b.getErr
	}
	if value, ok := b.values[user]; ok {
		return value, nil
	}
	return "", errKeyringItemNotFound
}

func (b *fakeKeyringBackend) Delete(service, user string) error {
	if b.deleteErr != nil {
		return b.deleteErr
	}
	delete(b.values, user)
	return nil
}

func TestKeyringStoreFallsBackToLessSecureFileWhenKeyringUnavailable(t *testing.T) {
	fallbackPath := filepath.Join(t.TempDir(), "tick", "auth-fallback.json")
	backend := &fakeKeyringBackend{
		setErr: errors.New("secret service not available"),
		getErr: errors.New("secret service not available"),
	}
	store := KeyringStore{
		Backend: backend,
		FallbackPath: func() (string, error) {
			return fallbackPath, nil
		},
	}

	token := Token{
		AccessToken:  "access-1",
		RefreshToken: "refresh-1",
		TokenType:    "Bearer",
	}
	if err := store.SaveToken(token); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}
	if err := store.SaveClientSecret("secret-1"); err != nil {
		t.Fatalf("SaveClientSecret() error = %v", err)
	}

	loadedToken, err := store.LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() error = %v", err)
	}
	if loadedToken.AccessToken != "access-1" {
		t.Fatalf("AccessToken = %q, want access-1", loadedToken.AccessToken)
	}
	loadedSecret, err := store.LoadClientSecret()
	if err != nil {
		t.Fatalf("LoadClientSecret() error = %v", err)
	}
	if loadedSecret != "secret-1" {
		t.Fatalf("client secret = %q, want secret-1", loadedSecret)
	}

	info, err := os.Stat(fallbackPath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("permissions = %#o, want 0600", got)
	}
	data, err := os.ReadFile(fallbackPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "less-secure-file-fallback") {
		t.Fatalf("fallback file = %q, want less-secure-file-fallback marker", string(data))
	}
}

func TestKeyringStoreLoadTokenReturnsClearErrorWhenKeyringUnavailableAndFallbackMissing(t *testing.T) {
	fallbackPath := filepath.Join(t.TempDir(), "tick", "auth-fallback.json")
	store := KeyringStore{
		Backend: &fakeKeyringBackend{
			getErr: errors.New("dbus-launch: no secret service"),
		},
		FallbackPath: func() (string, error) {
			return fallbackPath, nil
		},
	}

	_, err := store.LoadToken()
	if err == nil {
		t.Fatal("LoadToken() error = nil, want non-nil")
	}
	if strings.Contains(err.Error(), "dbus-launch") {
		t.Fatalf("error = %q, want sanitized message", err)
	}
	if !strings.Contains(err.Error(), "less-secure fallback file") {
		t.Fatalf("error = %q, want fallback guidance", err)
	}
	if !strings.Contains(err.Error(), fallbackPath) {
		t.Fatalf("error = %q, want fallback path", err)
	}
}
