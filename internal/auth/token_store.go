package auth

import (
	"encoding/json"
	"errors"

	"github.com/jeely/ticktick-cli/internal/domain"
	"github.com/zalando/go-keyring"
)

const (
	keyringService  = "tick"
	tokenKey        = "oauth-token"
	clientSecretKey = "oauth-client-secret"
)

var ErrNotAuthenticated = domain.ErrNotAuthenticated

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

type KeyringStore struct{}

func (KeyringStore) SaveToken(token Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return keyring.Set(keyringService, tokenKey, string(data))
}

func (KeyringStore) LoadToken() (Token, error) {
	value, err := keyring.Get(keyringService, tokenKey)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return Token{}, ErrNotAuthenticated
		}
		return Token{}, err
	}
	var token Token
	if err := json.Unmarshal([]byte(value), &token); err != nil {
		return Token{}, err
	}
	return token, nil
}

func (KeyringStore) DeleteToken() error {
	if err := keyring.Delete(keyringService, tokenKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return err
	}
	return nil
}

func (KeyringStore) SaveClientSecret(secret string) error {
	return keyring.Set(keyringService, clientSecretKey, secret)
}

func (KeyringStore) LoadClientSecret() (string, error) {
	value, err := keyring.Get(keyringService, clientSecretKey)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrNotAuthenticated
		}
		return "", err
	}
	return value, nil
}

func (KeyringStore) DeleteClientSecret() error {
	if err := keyring.Delete(keyringService, clientSecretKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return err
	}
	return nil
}
