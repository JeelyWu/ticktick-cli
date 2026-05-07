package auth

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileStoreSaveAndLoadToken(t *testing.T) {
	store := FileStore{Path: filepath.Join(t.TempDir(), "tick", "auth.json")}

	token := Token{
		AccessToken:  "access-1",
		RefreshToken: "refresh-1",
		TokenType:    "Bearer",
	}
	if err := store.SaveToken(token); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	loaded, err := store.LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() error = %v", err)
	}
	if loaded.AccessToken != "access-1" {
		t.Fatalf("AccessToken = %q, want access-1", loaded.AccessToken)
	}
	if loaded.ExpiresAtUnix != 0 {
		t.Fatalf("ExpiresAtUnix = %d, want 0", loaded.ExpiresAtUnix)
	}
}

func TestFileStoreSaveAndLoadClientSecret(t *testing.T) {
	store := FileStore{Path: filepath.Join(t.TempDir(), "tick", "auth.json")}

	if err := store.SaveClientSecret("secret-1"); err != nil {
		t.Fatalf("SaveClientSecret() error = %v", err)
	}

	loaded, err := store.LoadClientSecret()
	if err != nil {
		t.Fatalf("LoadClientSecret() error = %v", err)
	}
	if loaded != "secret-1" {
		t.Fatalf("client secret = %q, want secret-1", loaded)
	}
}

func TestFileStorePreservesTokenExpiryMetadata(t *testing.T) {
	store := FileStore{Path: filepath.Join(t.TempDir(), "tick", "auth.json")}

	token := Token{
		AccessToken:   "access-1",
		RefreshToken:  "refresh-1",
		ExpiresIn:     3600,
		CreatedAtUnix: 1_776_351_966,
		ExpiresAtUnix: 1_776_355_566,
	}
	if err := store.SaveToken(token); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}
	loaded, err := store.LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() error = %v", err)
	}
	if loaded.ExpiresIn != 3600 {
		t.Fatalf("ExpiresIn = %d, want 3600", loaded.ExpiresIn)
	}
	if loaded.CreatedAtUnix != 1_776_351_966 {
		t.Fatalf("CreatedAtUnix = %d, want 1776351966", loaded.CreatedAtUnix)
	}
	if loaded.ExpiresAtUnix != 1_776_355_566 {
		t.Fatalf("ExpiresAtUnix = %d, want 1776355566", loaded.ExpiresAtUnix)
	}
}

func TestFileStoreLoadTokenReturnsNotAuthenticatedWhenFileMissing(t *testing.T) {
	store := FileStore{Path: filepath.Join(t.TempDir(), "tick", "auth.json")}

	_, err := store.LoadToken()
	if !errors.Is(err, ErrNotAuthenticated) {
		t.Fatalf("LoadToken() error = %v, want ErrNotAuthenticated", err)
	}
}

func TestFileStoreLoadClientSecretReturnsNotAuthenticatedWhenFileMissing(t *testing.T) {
	store := FileStore{Path: filepath.Join(t.TempDir(), "tick", "auth.json")}

	_, err := store.LoadClientSecret()
	if !errors.Is(err, ErrNotAuthenticated) {
		t.Fatalf("LoadClientSecret() error = %v, want ErrNotAuthenticated", err)
	}
}

func TestFileStoreDeleteToken(t *testing.T) {
	store := FileStore{Path: filepath.Join(t.TempDir(), "tick", "auth.json")}

	if err := store.SaveToken(Token{AccessToken: "access-1"}); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}
	if err := store.DeleteToken(); err != nil {
		t.Fatalf("DeleteToken() error = %v", err)
	}
	_, err := store.LoadToken()
	if !errors.Is(err, ErrNotAuthenticated) {
		t.Fatalf("LoadToken() after delete error = %v, want ErrNotAuthenticated", err)
	}
}

func TestFileStoreDeleteClientSecret(t *testing.T) {
	store := FileStore{Path: filepath.Join(t.TempDir(), "tick", "auth.json")}

	if err := store.SaveClientSecret("secret-1"); err != nil {
		t.Fatalf("SaveClientSecret() error = %v", err)
	}
	if err := store.DeleteClientSecret(); err != nil {
		t.Fatalf("DeleteClientSecret() error = %v", err)
	}
	_, err := store.LoadClientSecret()
	if !errors.Is(err, ErrNotAuthenticated) {
		t.Fatalf("LoadClientSecret() after delete error = %v, want ErrNotAuthenticated", err)
	}
}

func TestFileStoreDeleteRemovesFileWhenEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tick", "auth.json")
	store := FileStore{Path: path}

	if err := store.SaveToken(Token{AccessToken: "access-1"}); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}
	if err := store.SaveClientSecret("secret-1"); err != nil {
		t.Fatalf("SaveClientSecret() error = %v", err)
	}
	if err := store.DeleteToken(); err != nil {
		t.Fatalf("DeleteToken() error = %v", err)
	}
	if err := store.DeleteClientSecret(); err != nil {
		t.Fatalf("DeleteClientSecret() error = %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Stat() error = %v, want not exist", err)
	}
}

func TestFileStoreFileHasRestrictedPermissions(t *testing.T) {
	store := FileStore{Path: filepath.Join(t.TempDir(), "tick", "auth.json")}

	if err := store.SaveToken(Token{AccessToken: "access-1"}); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	info, err := os.Stat(store.Path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("permissions = %#o, want 0600", got)
	}
}

func TestFileStoreRejectsInsecureDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tick", "auth.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	store := FileStore{Path: path}

	err := store.SaveToken(Token{AccessToken: "access-1"})
	if err == nil {
		t.Fatal("SaveToken() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "0700") {
		t.Fatalf("error = %q, want 0700 guidance", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Stat() error = %v, want not exist", err)
	}
}

func TestFileStoreTokenAndSecretCoexist(t *testing.T) {
	store := FileStore{Path: filepath.Join(t.TempDir(), "tick", "auth.json")}

	if err := store.SaveToken(Token{AccessToken: "access-1"}); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}
	if err := store.SaveClientSecret("secret-1"); err != nil {
		t.Fatalf("SaveClientSecret() error = %v", err)
	}

	token, err := store.LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() error = %v", err)
	}
	if token.AccessToken != "access-1" {
		t.Fatalf("AccessToken = %q, want access-1", token.AccessToken)
	}

	secret, err := store.LoadClientSecret()
	if err != nil {
		t.Fatalf("LoadClientSecret() error = %v", err)
	}
	if secret != "secret-1" {
		t.Fatalf("client secret = %q, want secret-1", secret)
	}
}

func TestFileStoreDeleteTokenPreservesSecret(t *testing.T) {
	store := FileStore{Path: filepath.Join(t.TempDir(), "tick", "auth.json")}

	if err := store.SaveToken(Token{AccessToken: "access-1"}); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}
	if err := store.SaveClientSecret("secret-1"); err != nil {
		t.Fatalf("SaveClientSecret() error = %v", err)
	}
	if err := store.DeleteToken(); err != nil {
		t.Fatalf("DeleteToken() error = %v", err)
	}

	_, err := store.LoadToken()
	if !errors.Is(err, ErrNotAuthenticated) {
		t.Fatalf("LoadToken() error = %v, want ErrNotAuthenticated", err)
	}

	secret, err := store.LoadClientSecret()
	if err != nil {
		t.Fatalf("LoadClientSecret() error = %v", err)
	}
	if secret != "secret-1" {
		t.Fatalf("client secret = %q, want secret-1", secret)
	}
}
