package auth

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
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
	Authenticated    bool
	HasRefresh       bool
	ExpiryKnown      bool
	ExpiresAtUnix    int64
	ExpiresInSeconds int64
	Expired          bool
}

type Service struct {
	AuthorizeURL string
	ClientID     string
	Exchanger    Exchanger
	Store        TokenStore
	Browser      Browser
	In           io.Reader
	Out          io.Writer
	StateSource  func() string
	Now          func() time.Time
	RefreshSkew  time.Duration
	Listen       func(network, address string) (net.Listener, error)
}

type fallbackPathReporter interface {
	ActiveFallbackPath() (string, bool, error)
}

func (s Service) Login(ctx context.Context, in LoginInput) (Token, error) {
	if in.ClientID == "" || in.ClientSecret == "" || in.RedirectURL == "" {
		return Token{}, errors.New("client-id, client-secret, and redirect-url are required")
	}

	state := s.state()
	authorizeURL := BuildAuthorizeURL(s.AuthorizeURL, OAuthConfig{
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
	response, err := s.authorizationResponse(ctx, in.RedirectURL)
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
		rollbackErr := s.Store.DeleteClientSecret()
		if rollbackErr != nil && !errors.Is(rollbackErr, ErrNotAuthenticated) {
			return Token{}, fmt.Errorf("%w; rollback client secret: %v", err, rollbackErr)
		}
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
	status := Status{
		Authenticated: token.AccessToken != "" || token.RefreshToken != "",
		HasRefresh:    token.RefreshToken != "",
	}
	if token.HasExpiry() {
		now := s.now()
		status.ExpiryKnown = true
		status.ExpiresAtUnix = token.ExpiresAtUnix
		status.Expired = !now.Before(token.ExpiresAt())
		if !status.Expired {
			status.ExpiresInSeconds = int64(token.ExpiresAt().Sub(now).Seconds())
		}
	}
	return status, nil
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
	if token.AccessToken != "" && !token.NeedsRefresh(s.now(), s.refreshSkew()) {
		return token.AccessToken, nil
	}
	if token.RefreshToken == "" {
		if token.AccessToken == "" {
			return "", ErrNotAuthenticated
		}
		return "", errors.New("stored access token expired and no refresh token is available; run `tick auth login`")
	}
	if s.ClientID == "" {
		return "", errors.New("oauth client-id is required to refresh tokens; run `tick auth login` again or set oauth.client_id")
	}
	clientSecret, err := s.Store.LoadClientSecret()
	if err != nil {
		return "", err
	}
	refreshed, err := s.Exchanger.RefreshToken(ctx, OAuthConfig{
		ClientID:     s.ClientID,
		ClientSecret: clientSecret,
	}, token.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("refreshing access token failed: %w; run `tick auth login`", err)
	}
	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = token.RefreshToken
	}
	if err := s.Store.SaveToken(refreshed); err != nil {
		return "", err
	}
	return refreshed.AccessToken, nil
}

func (s Service) authorizationResponse(ctx context.Context, redirectURL string) (string, error) {
	if receiver, err := s.startCallbackReceiver(ctx, redirectURL); err == nil {
		_, _ = fmt.Fprintf(s.Out, "Waiting for browser callback on %s or paste the full callback URL: ", redirectURL)
		return readAuthorizationResponseEither(ctx, s.In, receiver)
	} else if !errors.Is(err, errAutomaticCallbackUnsupported) {
		_, _ = fmt.Fprintf(s.Out, "Automatic callback unavailable (%v). Falling back to manual paste.\n", err)
	}

	_, _ = fmt.Fprint(s.Out, "Paste the full callback URL: ")
	return readAuthorizationResponse(s.In)
}

func (s Service) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func (s Service) refreshSkew() time.Duration {
	if s.RefreshSkew > 0 {
		return s.RefreshSkew
	}
	return time.Minute
}

var errAutomaticCallbackUnsupported = errors.New("automatic callback unsupported for redirect URL")

func (s Service) startCallbackReceiver(ctx context.Context, redirectURL string) (<-chan string, error) {
	callbackURL, err := url.Parse(redirectURL)
	if err != nil {
		return nil, errAutomaticCallbackUnsupported
	}
	if callbackURL.Scheme != "http" {
		return nil, errAutomaticCallbackUnsupported
	}
	host := callbackURL.Hostname()
	if host != "127.0.0.1" && host != "localhost" && host != "::1" {
		return nil, errAutomaticCallbackUnsupported
	}

	listen := s.Listen
	if listen == nil {
		listen = net.Listen
	}
	listener, err := listen("tcp", callbackURL.Host)
	if err != nil {
		return nil, err
	}

	responseCh := make(chan string, 1)
	server := &http.Server{}
	var once sync.Once
	send := func(value string) {
		once.Do(func() {
			responseCh <- value
			close(responseCh)
		})
	}

	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != callbackURL.Path {
			http.NotFound(w, r)
			return
		}
		fullURL := callbackURL.Scheme + "://" + callbackURL.Host + r.URL.RequestURI()
		_, _ = io.WriteString(w, "Authorization received. You can return to the terminal.")
		send(fullURL)
		go func() {
			_ = server.Shutdown(context.Background())
		}()
	})

	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			send("")
		}
	}()
	return responseCh, nil
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

func readAuthorizationResponseEither(ctx context.Context, r io.Reader, callback <-chan string) (string, error) {
	type result struct {
		response string
		err      error
	}
	manualCh := make(chan result, 1)
	go func() {
		response, err := readAuthorizationResponse(r)
		manualCh <- result{response: response, err: err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case response, ok := <-callback:
		if ok && response != "" {
			return response, nil
		}
		return readAuthorizationResponse(r)
	case result := <-manualCh:
		return result.response, result.err
	}
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
