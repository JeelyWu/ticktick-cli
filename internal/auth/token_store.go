package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jeely/ticktick-cli/internal/domain"
)

const authFileName = "auth.json"

var ErrNotAuthenticated = domain.ErrNotAuthenticated

type credentials struct {
	Token        *Token `json:"token,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

type Token struct {
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	TokenType     string `json:"token_type"`
	Scope         string `json:"scope"`
	ExpiresIn     int64  `json:"expires_in,omitempty"`
	CreatedAtUnix int64  `json:"created_at,omitempty"`
	ExpiresAtUnix int64  `json:"expires_at,omitempty"`
}

func (t Token) HasExpiry() bool {
	return t.ExpiresAtUnix > 0
}

func (t Token) ExpiresAt() time.Time {
	if t.ExpiresAtUnix == 0 {
		return time.Time{}
	}
	return time.Unix(t.ExpiresAtUnix, 0).UTC()
}

func (t Token) NeedsRefresh(now time.Time, skew time.Duration) bool {
	if !t.HasExpiry() {
		return false
	}
	return !now.UTC().Before(t.ExpiresAt().Add(-skew))
}

func (t Token) withExpiry(now time.Time) Token {
	if t.ExpiresIn <= 0 {
		return t
	}
	now = now.UTC()
	t.CreatedAtUnix = now.Unix()
	t.ExpiresAtUnix = now.Add(time.Duration(t.ExpiresIn) * time.Second).Unix()
	return t
}

type TokenStore interface {
	SaveToken(Token) error
	LoadToken() (Token, error)
	DeleteToken() error
	SaveClientSecret(string) error
	LoadClientSecret() (string, error)
	DeleteClientSecret() error
}

type FileStore struct {
	Path string
}

func (s FileStore) SaveToken(token Token) error {
	return s.update(func(c *credentials) {
		copyToken := token
		c.Token = &copyToken
	})
}

func (s FileStore) LoadToken() (Token, error) {
	creds, err := s.read()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Token{}, ErrNotAuthenticated
		}
		return Token{}, err
	}
	if creds.Token == nil {
		return Token{}, ErrNotAuthenticated
	}
	return *creds.Token, nil
}

func (s FileStore) DeleteToken() error {
	return s.update(func(c *credentials) {
		c.Token = nil
	})
}

func (s FileStore) SaveClientSecret(secret string) error {
	return s.update(func(c *credentials) {
		c.ClientSecret = secret
	})
}

func (s FileStore) LoadClientSecret() (string, error) {
	creds, err := s.read()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrNotAuthenticated
		}
		return "", err
	}
	if creds.ClientSecret == "" {
		return "", ErrNotAuthenticated
	}
	return creds.ClientSecret, nil
}

func (s FileStore) DeleteClientSecret() error {
	return s.update(func(c *credentials) {
		c.ClientSecret = ""
	})
}

func (s FileStore) filePath() string {
	if s.Path != "" {
		return s.Path
	}
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "tick", authFileName)
}

func (s FileStore) read() (credentials, error) {
	path := s.filePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return credentials{}, err
	}
	var creds credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return credentials{}, err
	}
	return creds, nil
}

func (s FileStore) update(update func(*credentials)) error {
	path := s.filePath()

	creds := credentials{}
	existing, err := s.read()
	if err == nil {
		creds = existing
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	update(&creds)

	if creds.Token == nil && creds.ClientSecret == "" {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}

	return writeAuthFile(path, creds)
}

func writeAuthFile(path string, creds credentials) error {
	dir := filepath.Dir(path)
	if err := ensurePrivateDir(dir); err != nil {
		return err
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	tempFile, err := os.CreateTemp(dir, "auth-*.tmp")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	success := false
	defer func() {
		if !success {
			_ = os.Remove(tempPath)
		}
	}()
	if supportsPOSIXPrivatePerms() {
		if err := tempFile.Chmod(0o600); err != nil {
			_ = tempFile.Close()
			return err
		}
	}
	if _, err := tempFile.Write(data); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	success = true
	return nil
}

func ensurePrivateDir(path string) error {
	info, err := os.Lstat(path)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("auth directory %s must not be a symlink", path)
		}
		if !info.IsDir() {
			return fmt.Errorf("auth directory %s is not a directory", path)
		}
		if supportsPOSIXPrivatePerms() && info.Mode().Perm() != 0o700 {
			return fmt.Errorf("auth directory %s must have permissions 0700", path)
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(path, 0o700); err != nil {
		return err
	}
	return nil
}

func supportsPOSIXPrivatePerms() bool {
	return runtime.GOOS != "windows"
}
