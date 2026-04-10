package auth

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

type stubBrowser struct {
	err  error
	urls []string
}

func (b *stubBrowser) OpenURL(url string) error {
	b.urls = append(b.urls, url)
	return b.err
}

type memoryTokenStore struct {
	token             Token
	clientSecret      string
	saveTokenCalls    int
	saveSecretCalls   int
	deleteTokenCalls  int
	deleteSecretCalls int
}

func (s *memoryTokenStore) SaveToken(token Token) error {
	s.saveTokenCalls++
	s.token = token
	return nil
}

func (s *memoryTokenStore) LoadToken() (Token, error) {
	if s.token.AccessToken == "" && s.token.RefreshToken == "" && s.token.TokenType == "" && s.token.Scope == "" {
		return Token{}, ErrNotAuthenticated
	}
	return s.token, nil
}

func (s *memoryTokenStore) DeleteToken() error {
	s.deleteTokenCalls++
	s.token = Token{}
	return nil
}

func (s *memoryTokenStore) SaveClientSecret(secret string) error {
	s.saveSecretCalls++
	s.clientSecret = secret
	return nil
}

func (s *memoryTokenStore) LoadClientSecret() (string, error) {
	if s.clientSecret == "" {
		return "", ErrNotAuthenticated
	}
	return s.clientSecret, nil
}

func (s *memoryTokenStore) DeleteClientSecret() error {
	s.deleteSecretCalls++
	s.clientSecret = ""
	return nil
}

type deleteErrorStore struct {
	tokenErr  error
	secretErr error
}

func (s deleteErrorStore) SaveToken(Token) error             { return nil }
func (s deleteErrorStore) LoadToken() (Token, error)         { return Token{}, ErrNotAuthenticated }
func (s deleteErrorStore) DeleteToken() error                { return s.tokenErr }
func (s deleteErrorStore) SaveClientSecret(string) error     { return nil }
func (s deleteErrorStore) LoadClientSecret() (string, error) { return "", ErrNotAuthenticated }
func (s deleteErrorStore) DeleteClientSecret() error         { return s.secretErr }

func TestServiceLoginContinuesWhenBrowserOpenFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"access-1","refresh_token":"refresh-1","token_type":"Bearer"}`))
	}))
	defer server.Close()

	state := "state-1"
	callbackURL := "http://localhost:14573/callback?code=code-1&state=" + state
	store := &memoryTokenStore{}
	browser := &stubBrowser{err: errors.New("browser unavailable")}
	out := &bytes.Buffer{}

	token, err := Service{
		Exchanger: Exchanger{
			HTTPClient: server.Client(),
			TokenURL:   server.URL,
		},
		Store:       store,
		Browser:     browser,
		In:          strings.NewReader(callbackURL + "\n"),
		Out:         out,
		StateSource: func() string { return state },
	}.Login(context.Background(), LoginInput{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  "http://localhost:14573/callback",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if token.AccessToken != "access-1" {
		t.Fatalf("AccessToken = %q, want access-1", token.AccessToken)
	}
	if store.saveSecretCalls != 1 {
		t.Fatalf("SaveClientSecret() calls = %d, want 1", store.saveSecretCalls)
	}
	if store.saveTokenCalls != 1 {
		t.Fatalf("SaveToken() calls = %d, want 1", store.saveTokenCalls)
	}
	if len(browser.urls) != 1 {
		t.Fatalf("OpenURL() calls = %d, want 1", len(browser.urls))
	}
	if !strings.Contains(out.String(), "Could not open browser automatically") {
		t.Fatalf("output = %q, want browser fallback warning", out.String())
	}
	if !strings.Contains(out.String(), "https://ticktick.com/oauth/authorize?") {
		t.Fatalf("output = %q, want authorize URL", out.String())
	}
}

func TestServiceLoginWarnsWhenLessSecureFallbackFileIsUsed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"access-1","refresh_token":"refresh-1","token_type":"Bearer"}`))
	}))
	defer server.Close()

	state := "state-1"
	fallbackPath := filepath.Join(t.TempDir(), "tick", "auth-fallback", "auth-fallback.json")
	out := &bytes.Buffer{}

	_, err := Service{
		Exchanger: Exchanger{
			HTTPClient: server.Client(),
			TokenURL:   server.URL,
		},
		Store: KeyringStore{
			Backend: &fakeKeyringBackend{
				setErr: errors.New("secret service not available"),
			},
			FallbackPath: func() (string, error) {
				return fallbackPath, nil
			},
		},
		In:          strings.NewReader("http://localhost:14573/callback?code=code-1&state=" + state + "\n"),
		Out:         out,
		StateSource: func() string { return state },
	}.Login(context.Background(), LoginInput{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  "http://localhost:14573/callback",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if !strings.Contains(out.String(), "less-secure fallback file") {
		t.Fatalf("output = %q, want fallback warning", out.String())
	}
	if !strings.Contains(out.String(), fallbackPath) {
		t.Fatalf("output = %q, want fallback path", out.String())
	}
}

func TestServiceLoginRejectsMismatchedState(t *testing.T) {
	store := &memoryTokenStore{}

	_, err := Service{
		Store:       store,
		In:          strings.NewReader("http://localhost:14573/callback?code=code-1&state=wrong-state\n"),
		Out:         &bytes.Buffer{},
		StateSource: func() string { return "state-1" },
	}.Login(context.Background(), LoginInput{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  "http://localhost:14573/callback",
	})
	if err == nil {
		t.Fatal("Login() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "state mismatch") {
		t.Fatalf("error = %q, want state mismatch", err)
	}
	if store.saveTokenCalls != 0 {
		t.Fatalf("SaveToken() calls = %d, want 0", store.saveTokenCalls)
	}
}

func TestServiceLoginAcceptsCallbackURLWithoutTrailingNewline(t *testing.T) {
	var exchangedCode string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		exchangedCode = r.PostForm.Get("code")
		_, _ = w.Write([]byte(`{"access_token":"access-1","refresh_token":"refresh-1","token_type":"Bearer"}`))
	}))
	defer server.Close()

	state := "state-1"
	callbackURL := "http://localhost:14573/callback?code=code-1&state=" + state
	store := &memoryTokenStore{}

	token, err := Service{
		Exchanger: Exchanger{
			HTTPClient: server.Client(),
			TokenURL:   server.URL,
		},
		Store:       store,
		In:          strings.NewReader(callbackURL),
		Out:         &bytes.Buffer{},
		StateSource: func() string { return state },
	}.Login(context.Background(), LoginInput{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  "http://localhost:14573/callback",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if token.AccessToken != "access-1" {
		t.Fatalf("AccessToken = %q, want access-1", token.AccessToken)
	}
	if exchangedCode != "code-1" {
		t.Fatalf("exchanged code = %q, want code-1", exchangedCode)
	}
}

func TestBuildAuthorizeURLIncludesSuppliedState(t *testing.T) {
	got := BuildAuthorizeURL(OAuthConfig{
		ClientID:    "client-1",
		RedirectURL: "http://localhost:14573/callback",
	}, "state-1")

	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	if parsed.Query().Get("state") != "state-1" {
		t.Fatalf("state = %q, want state-1", parsed.Query().Get("state"))
	}
}

func TestServiceStatusTreatsFallbackGuidanceAsNotAuthenticated(t *testing.T) {
	service := Service{
		Store: KeyringStore{
			Backend: &fakeKeyringBackend{
				getErr: errors.New("dbus-launch: no secret service"),
			},
			FallbackPath: func() (string, error) {
				return filepath.Join(t.TempDir(), "tick", "auth-fallback", "auth-fallback.json"), nil
			},
		},
	}

	status, err := service.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Authenticated {
		t.Fatal("Authenticated = true, want false")
	}
}

func TestServiceLogoutDeletesTokenAndClientSecret(t *testing.T) {
	store := &memoryTokenStore{
		token:        Token{AccessToken: "access-1"},
		clientSecret: "secret-1",
	}

	service := Service{Store: store}
	if err := service.Logout(context.Background()); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if store.deleteTokenCalls != 1 {
		t.Fatalf("DeleteToken() calls = %d, want 1", store.deleteTokenCalls)
	}
	if store.deleteSecretCalls != 1 {
		t.Fatalf("DeleteClientSecret() calls = %d, want 1", store.deleteSecretCalls)
	}
}

func TestServiceLogoutIgnoresNotAuthenticatedDeletes(t *testing.T) {
	service := Service{Store: deleteErrorStore{
		tokenErr:  ErrNotAuthenticated,
		secretErr: ErrNotAuthenticated,
	}}
	if err := service.Logout(context.Background()); err != nil {
		t.Fatalf("Logout() error = %v, want nil", err)
	}
}
