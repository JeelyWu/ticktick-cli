package auth

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
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
	StateSource func() string
}

type fallbackPathReporter interface {
	ActiveFallbackPath() (string, bool, error)
}

func (s Service) Login(ctx context.Context, in LoginInput) (Token, error) {
	if in.ClientID == "" || in.ClientSecret == "" || in.RedirectURL == "" {
		return Token{}, errors.New("client-id, client-secret, and redirect-url are required")
	}

	state := s.state()
	authorizeURL := BuildAuthorizeURL(OAuthConfig{
		ClientID:    in.ClientID,
		RedirectURL: in.RedirectURL,
	}, state)
	if s.Browser != nil {
		if err := s.Browser.OpenURL(authorizeURL); err != nil {
			_, _ = fmt.Fprintln(s.Out, "Could not open browser automatically. Open this URL manually:")
			_, _ = fmt.Fprintln(s.Out, authorizeURL)
		} else {
			_, _ = fmt.Fprintln(s.Out, "Open this URL in your browser if it did not open automatically:")
			_, _ = fmt.Fprintln(s.Out, authorizeURL)
		}
	} else {
		_, _ = fmt.Fprintln(s.Out, "Open this URL in your browser if it did not open automatically:")
		_, _ = fmt.Fprintln(s.Out, authorizeURL)
	}
	_, _ = fmt.Fprint(s.Out, "Paste the full callback URL: ")

	response, err := readAuthorizationResponse(s.In)
	if err != nil {
		return Token{}, err
	}
	code, err := parseAuthorizationCode(response, in.RedirectURL, state)
	if err != nil {
		return Token{}, err
	}
	token, err := s.Exchanger.ExchangeCode(ctx, OAuthConfig{
		ClientID:     in.ClientID,
		ClientSecret: in.ClientSecret,
		RedirectURL:  in.RedirectURL,
	}, code)
	if err != nil {
		return Token{}, err
	}
	if err := s.Store.SaveClientSecret(in.ClientSecret); err != nil {
		return Token{}, err
	}
	if err := s.Store.SaveToken(token); err != nil {
		return Token{}, err
	}
	if reporter, ok := s.Store.(fallbackPathReporter); ok {
		path, active, err := reporter.ActiveFallbackPath()
		if err != nil {
			return Token{}, err
		}
		if active {
			_, _ = fmt.Fprintf(s.Out, "Warning: system keyring unavailable; credentials were stored in the less-secure fallback file at %s\n", path)
		}
	}
	return token, nil
}

func (s Service) Status(ctx context.Context) (Status, error) {
	token, err := s.Store.LoadToken()
	if err != nil {
		if errors.Is(err, ErrNotAuthenticated) {
			return Status{}, nil
		}
		var guidanceErr fallbackLoginRequiredError
		if errors.As(err, &guidanceErr) {
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
	tokenErr := s.Store.DeleteToken()
	secretErr := s.Store.DeleteClientSecret()

	if tokenErr != nil && !errors.Is(tokenErr, ErrNotAuthenticated) {
		return tokenErr
	}
	if secretErr != nil && !errors.Is(secretErr, ErrNotAuthenticated) {
		return secretErr
	}
	return nil
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

func (s Service) state() string {
	if s.StateSource != nil {
		return s.StateSource()
	}
	return randomState()
}

func readAuthorizationResponse(r io.Reader) (string, error) {
	response, err := bufio.NewReader(r).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	response = strings.TrimSpace(response)
	if response == "" {
		if err != nil {
			return "", err
		}
		return "", errors.New("authorization response is required")
	}
	return response, nil
}

func parseAuthorizationCode(response string, redirectURL string, expectedState string) (string, error) {
	callbackURL, err := url.Parse(strings.TrimSpace(response))
	if err != nil || callbackURL.Scheme == "" || callbackURL.Host == "" {
		return "", errors.New("authorization response must be the full callback URL")
	}
	expected, err := url.Parse(redirectURL)
	if err != nil {
		return "", err
	}
	if callbackURL.Scheme != expected.Scheme || callbackURL.Host != expected.Host || callbackURL.Path != expected.Path {
		return "", errors.New("authorization response redirect does not match redirect-url")
	}

	values := callbackURL.Query()
	expectedValues := expected.Query()
	for key, expectedValue := range expectedValues {
		if !equalQueryValues(values[key], expectedValue) {
			return "", errors.New("authorization response redirect does not match redirect-url")
		}
	}
	if authError := values.Get("error"); authError != "" {
		return "", fmt.Errorf("oauth authorization failed: %s", authError)
	}
	if values.Get("state") != expectedState {
		return "", errors.New("authorization response state mismatch")
	}

	code := values.Get("code")
	if code == "" {
		return "", errors.New("authorization response missing code")
	}
	return code, nil
}

func equalQueryValues(got []string, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
