package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jeely/ticktick-cli/internal/domain"
	"github.com/zalando/go-keyring"
)

const (
	keyringService       = "tick"
	tokenKey             = "oauth-token"
	clientSecretKey      = "oauth-client-secret"
	fallbackDirName      = "auth-fallback"
	fallbackFileName     = "auth-fallback.json"
	fallbackStorageLabel = "less-secure-file-fallback"
	tempFallbackDirName  = "tick-auth-fallback"
)

var ErrNotAuthenticated = domain.ErrNotAuthenticated
var errKeyringItemNotFound = keyring.ErrNotFound

type fallbackLoginRequiredError struct {
	path string
}

func (e fallbackLoginRequiredError) Error() string {
	return fmt.Sprintf("system keyring unavailable; run `tick auth login` to use the less-secure fallback file at %s", e.path)
}

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
		s.cleanupFallbackBestEffort(func(credentials *fallbackCredentials) {
			credentials.Token = nil
		})
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
			credentials, found, err := s.loadFallbackIfPresent()
			if err != nil {
				return Token{}, err
			}
			if found && credentials.Token != nil {
				return *credentials.Token, nil
			}
			return Token{}, ErrNotAuthenticated
		}
		if !isKeyringUnavailable(err) {
			return Token{}, err
		}
		credentials, found, err := s.loadFallbackIfPresent()
		if err != nil {
			return Token{}, err
		}
		if found && credentials.Token != nil {
			return *credentials.Token, nil
		}
		return Token{}, s.fallbackGuidanceError()
	}
	var token Token
	if err := json.Unmarshal([]byte(value), &token); err != nil {
		return Token{}, err
	}
	return token, nil
}

func (s KeyringStore) DeleteToken() error {
	if err := s.backend().Delete(keyringService, tokenKey); err != nil {
		if !errors.Is(err, errKeyringItemNotFound) && !isKeyringUnavailable(err) {
			return err
		}
		return s.updateFallback(func(credentials *fallbackCredentials) {
			credentials.Token = nil
		})
	}
	s.cleanupFallbackBestEffort(func(credentials *fallbackCredentials) {
		credentials.Token = nil
	})
	return nil
}

func (s KeyringStore) SaveClientSecret(secret string) error {
	err := s.backend().Set(keyringService, clientSecretKey, secret)
	if err == nil {
		s.cleanupFallbackBestEffort(func(credentials *fallbackCredentials) {
			credentials.ClientSecret = ""
		})
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
			credentials, found, err := s.loadFallbackIfPresent()
			if err != nil {
				return "", err
			}
			if found && credentials.ClientSecret != "" {
				return credentials.ClientSecret, nil
			}
			return "", ErrNotAuthenticated
		}
		if !isKeyringUnavailable(err) {
			return "", err
		}
		credentials, found, err := s.loadFallbackIfPresent()
		if err != nil {
			return "", err
		}
		if found && credentials.ClientSecret != "" {
			return credentials.ClientSecret, nil
		}
		return "", s.fallbackGuidanceError()
	}
	return value, nil
}

func (s KeyringStore) DeleteClientSecret() error {
	if err := s.backend().Delete(keyringService, clientSecretKey); err != nil {
		if !errors.Is(err, errKeyringItemNotFound) && !isKeyringUnavailable(err) {
			return err
		}
		return s.updateFallback(func(credentials *fallbackCredentials) {
			credentials.ClientSecret = ""
		})
	}
	s.cleanupFallbackBestEffort(func(credentials *fallbackCredentials) {
		credentials.ClientSecret = ""
	})
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
		return filepath.Join(tempFallbackDir(), fallbackFileName), nil
	}
	return filepath.Join(dir, "tick", fallbackDirName, fallbackFileName), nil
}

func (s KeyringStore) loadFallbackIfPresent() (fallbackCredentials, bool, error) {
	path, err := s.fallbackPath()
	if err != nil {
		return fallbackCredentials{}, false, err
	}
	if err := ensurePrivateFallbackDir(filepath.Dir(path), false); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fallbackCredentials{}, false, nil
		}
		return fallbackCredentials{}, false, err
	}
	credentials, err := readFallbackFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fallbackCredentials{}, false, nil
		}
		return fallbackCredentials{}, false, err
	}
	return credentials, true, nil
}

func (s KeyringStore) fallbackGuidanceError() error {
	path, err := s.fallbackPath()
	if err != nil {
		return err
	}
	return fallbackLoginRequiredError{path: path}
}

func (s KeyringStore) updateFallback(update func(*fallbackCredentials)) error {
	path, err := s.fallbackPath()
	if err != nil {
		return err
	}
	if err := ensurePrivateFallbackDir(filepath.Dir(path), true); err != nil {
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

func (s KeyringStore) cleanupFallbackBestEffort(update func(*fallbackCredentials)) {
	path, err := s.fallbackPath()
	if err != nil {
		return
	}
	if err := ensurePrivateFallbackDir(filepath.Dir(path), false); err != nil {
		return
	}

	credentials, err := readFallbackFile(path)
	if err != nil {
		return
	}

	update(&credentials)
	if credentials.Token == nil && credentials.ClientSecret == "" {
		_ = os.Remove(path)
		return
	}
	credentials.Storage = fallbackStorageLabel
	_ = writeFallbackFile(path, credentials)
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
	dir := filepath.Dir(path)
	if err := ensurePrivateFallbackDir(dir, true); err != nil {
		return err
	}
	data, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return err
	}
	tempFile, err := os.CreateTemp(dir, "auth-fallback-*.tmp")
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

func isKeyringUnavailable(err error) bool {
	if err == nil || errors.Is(err, errKeyringItemNotFound) {
		return false
	}
	if errors.Is(err, keyring.ErrUnsupportedPlatform) {
		return true
	}

	message := strings.ToLower(err.Error())
	hasUnavailableSignal := strings.Contains(message, "not available") ||
		strings.Contains(message, "unavailable")
	hasPortableBackendSignal := strings.Contains(message, "credential manager") ||
		strings.Contains(message, "credentials manager") ||
		strings.Contains(message, "keychain") ||
		strings.Contains(message, "keyring")
	return strings.Contains(message, "org.freedesktop.secrets") ||
		strings.Contains(message, "secret service not available") ||
		strings.Contains(message, "no secret service") ||
		strings.Contains(message, "dbus-launch") ||
		(hasUnavailableSignal && hasPortableBackendSignal)
}

func ensurePrivateFallbackDir(path string, create bool) error {
	info, err := os.Lstat(path)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("fallback directory %s must not be a symlink", path)
		}
		if !info.IsDir() {
			return fmt.Errorf("fallback directory %s is not a directory", path)
		}
		if supportsPOSIXPrivatePerms() && info.Mode().Perm() != 0o700 {
			return fmt.Errorf("fallback directory %s must have permissions 0700", path)
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) || !create {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.Mkdir(path, 0o700); err != nil {
		if errors.Is(err, os.ErrExist) {
			return ensurePrivateFallbackDir(path, false)
		}
		return err
	}
	return nil
}

func (s KeyringStore) ActiveFallbackPath() (string, bool, error) {
	path, err := s.fallbackPath()
	if err != nil {
		return "", false, err
	}
	if err := ensurePrivateFallbackDir(filepath.Dir(path), false); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	return path, true, nil
}

func tempFallbackDir() string {
	suffix := "default"
	if current, err := user.Current(); err == nil {
		switch {
		case current.Uid != "":
			suffix = current.Uid
		case current.Username != "":
			suffix = sanitizeFallbackSuffix(current.Username)
		}
	}
	return filepath.Join(os.TempDir(), tempFallbackDirName+"-"+suffix)
}

func sanitizeFallbackSuffix(value string) string {
	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-' || r == '_':
			builder.WriteRune(r)
		default:
			builder.WriteByte('-')
		}
	}
	if builder.Len() == 0 {
		return "default"
	}
	return builder.String()
}

func supportsPOSIXPrivatePerms() bool {
	return runtime.GOOS != "windows"
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
