package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jeely/ticktick-cli/internal/domain"
	"github.com/zalando/go-keyring"
)

const (
	keyringService       = "tick"
	tokenKey             = "oauth-token"
	clientSecretKey      = "oauth-client-secret"
	fallbackFileName     = "auth-fallback.json"
	fallbackStorageLabel = "less-secure-file-fallback"
)

var ErrNotAuthenticated = domain.ErrNotAuthenticated
var errKeyringItemNotFound = keyring.ErrNotFound

type keyringBackend interface {
	Set(service, user, value string) error
	Get(service, user string) (string, error)
	Delete(service, user string) error
}

type fallbackPathResolver func() (string, error)

type fallbackCredentials struct {
	Storage      string `json:"storage"`
	Token        *Token `json:"token,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

type TokenStore interface {
	SaveToken(Token) error
	LoadToken() (Token, error)
	DeleteToken() error
	SaveClientSecret(string) error
	LoadClientSecret() (string, error)
	DeleteClientSecret() error
}

type KeyringStore struct {
	Backend      keyringBackend
	FallbackPath fallbackPathResolver
}

func (s KeyringStore) SaveToken(token Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	err = s.backend().Set(keyringService, tokenKey, string(data))
	if err == nil {
		return nil
	}
	if !isKeyringUnavailable(err) {
		return err
	}
	return s.updateFallback(func(credentials *fallbackCredentials) {
		copyToken := token
		credentials.Token = &copyToken
	})
}

func (s KeyringStore) LoadToken() (Token, error) {
	value, err := s.backend().Get(keyringService, tokenKey)
	if err != nil {
		if errors.Is(err, errKeyringItemNotFound) {
			return Token{}, ErrNotAuthenticated
		}
		if !isKeyringUnavailable(err) {
			return Token{}, err
		}
		credentials, err := s.loadFallback()
		if err != nil {
			return Token{}, err
		}
		if credentials.Token == nil {
			return Token{}, ErrNotAuthenticated
		}
		return *credentials.Token, nil
	}
	var token Token
	if err := json.Unmarshal([]byte(value), &token); err != nil {
		return Token{}, err
	}
	return token, nil
}

func (s KeyringStore) DeleteToken() error {
	if err := s.backend().Delete(keyringService, tokenKey); err != nil {
		if errors.Is(err, errKeyringItemNotFound) {
			return nil
		}
		if !isKeyringUnavailable(err) {
			return err
		}
		return s.updateFallback(func(credentials *fallbackCredentials) {
			credentials.Token = nil
		})
	}
	return nil
}

func (s KeyringStore) SaveClientSecret(secret string) error {
	err := s.backend().Set(keyringService, clientSecretKey, secret)
	if err == nil {
		return nil
	}
	if !isKeyringUnavailable(err) {
		return err
	}
	return s.updateFallback(func(credentials *fallbackCredentials) {
		credentials.ClientSecret = secret
	})
}

func (s KeyringStore) LoadClientSecret() (string, error) {
	value, err := s.backend().Get(keyringService, clientSecretKey)
	if err != nil {
		if errors.Is(err, errKeyringItemNotFound) {
			return "", ErrNotAuthenticated
		}
		if !isKeyringUnavailable(err) {
			return "", err
		}
		credentials, err := s.loadFallback()
		if err != nil {
			return "", err
		}
		if credentials.ClientSecret == "" {
			return "", ErrNotAuthenticated
		}
		return credentials.ClientSecret, nil
	}
	return value, nil
}

func (s KeyringStore) DeleteClientSecret() error {
	if err := s.backend().Delete(keyringService, clientSecretKey); err != nil {
		if errors.Is(err, errKeyringItemNotFound) {
			return nil
		}
		if !isKeyringUnavailable(err) {
			return err
		}
		return s.updateFallback(func(credentials *fallbackCredentials) {
			credentials.ClientSecret = ""
		})
	}
	return nil
}

func (s KeyringStore) backend() keyringBackend {
	if s.Backend != nil {
		return s.Backend
	}
	return defaultKeyringBackend{}
}

func (s KeyringStore) fallbackPath() (string, error) {
	if s.FallbackPath != nil {
		return s.FallbackPath()
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("system keyring unavailable and less-secure fallback file path could not be resolved: %w", err)
	}
	return filepath.Join(dir, "tick", fallbackFileName), nil
}

func (s KeyringStore) loadFallback() (fallbackCredentials, error) {
	path, err := s.fallbackPath()
	if err != nil {
		return fallbackCredentials{}, err
	}
	credentials, err := readFallbackFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fallbackCredentials{}, fmt.Errorf("system keyring unavailable; run `tick auth login` to use the less-secure fallback file at %s", path)
		}
		return fallbackCredentials{}, err
	}
	return credentials, nil
}

func (s KeyringStore) updateFallback(update func(*fallbackCredentials)) error {
	path, err := s.fallbackPath()
	if err != nil {
		return err
	}

	credentials := fallbackCredentials{Storage: fallbackStorageLabel}
	existing, err := readFallbackFile(path)
	if err == nil {
		credentials = existing
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	update(&credentials)
	if credentials.Token == nil && credentials.ClientSecret == "" {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	credentials.Storage = fallbackStorageLabel
	return writeFallbackFile(path, credentials)
}

func readFallbackFile(path string) (fallbackCredentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallbackCredentials{}, err
	}
	var credentials fallbackCredentials
	if err := json.Unmarshal(data, &credentials); err != nil {
		return fallbackCredentials{}, err
	}
	return credentials, nil
}

func writeFallbackFile(path string, credentials fallbackCredentials) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func isKeyringUnavailable(err error) bool {
	return err != nil && !errors.Is(err, errKeyringItemNotFound)
}

type defaultKeyringBackend struct{}

func (defaultKeyringBackend) Set(service, user, value string) error {
	return keyring.Set(service, user, value)
}

func (defaultKeyringBackend) Get(service, user string) (string, error) {
	return keyring.Get(service, user)
}

func (defaultKeyringBackend) Delete(service, user string) error {
	return keyring.Delete(service, user)
}
