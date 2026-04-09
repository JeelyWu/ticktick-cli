package auth

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Browser interface {
	OpenURL(string) error
}

type LoginInput struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type Status struct {
	Authenticated bool
	HasRefresh    bool
}

type Service struct {
	Exchanger Exchanger
	Store     TokenStore
	Browser   Browser
	In        io.Reader
	Out       io.Writer
}

func (s Service) Login(ctx context.Context, in LoginInput) (Token, error) {
	if in.ClientID == "" || in.ClientSecret == "" || in.RedirectURL == "" {
		return Token{}, errors.New("client-id, client-secret, and redirect-url are required")
	}

	state := randomState()
	url := BuildAuthorizeURL(OAuthConfig{
		ClientID:    in.ClientID,
		RedirectURL: in.RedirectURL,
	}, state)
	if s.Browser != nil {
		if err := s.Browser.OpenURL(url); err != nil {
			return Token{}, err
		}
	}
	_, _ = fmt.Fprintln(s.Out, "Open this URL in your browser if it did not open automatically:")
	_, _ = fmt.Fprintln(s.Out, url)
	_, _ = fmt.Fprint(s.Out, "Paste the authorization code: ")

	code, err := bufio.NewReader(s.In).ReadString('\n')
	if err != nil {
		return Token{}, err
	}
	token, err := s.Exchanger.ExchangeCode(ctx, OAuthConfig{
		ClientID:     in.ClientID,
		ClientSecret: in.ClientSecret,
		RedirectURL:  in.RedirectURL,
	}, strings.TrimSpace(code))
	if err != nil {
		return Token{}, err
	}
	if err := s.Store.SaveClientSecret(in.ClientSecret); err != nil {
		return Token{}, err
	}
	if err := s.Store.SaveToken(token); err != nil {
		return Token{}, err
	}
	return token, nil
}

func (s Service) Status(ctx context.Context) (Status, error) {
	token, err := s.Store.LoadToken()
	if err != nil {
		if errors.Is(err, ErrNotAuthenticated) {
			return Status{}, nil
		}
		return Status{}, err
	}
	return Status{
		Authenticated: token.AccessToken != "",
		HasRefresh:    token.RefreshToken != "",
	}, nil
}

func (s Service) Logout(ctx context.Context) error {
	return s.Store.DeleteToken()
}

func (s Service) AccessToken(ctx context.Context) (string, error) {
	token, err := s.Store.LoadToken()
	if err != nil {
		return "", err
	}
	if token.AccessToken == "" {
		return "", ErrNotAuthenticated
	}
	return token.AccessToken, nil
}

func randomState() string {
	var buf [16]byte
	_, _ = rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}
