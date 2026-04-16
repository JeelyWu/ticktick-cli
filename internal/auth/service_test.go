package auth

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	saveTokenErr      error
}

func (s *memoryTokenStore) SaveToken(token Token) error {
	s.saveTokenCalls++
	if s.saveTokenErr != nil {
		return s.saveTokenErr
	}
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

func TestServiceLoginRollsBackClientSecretWhenSavingTokenFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"access-1","refresh_token":"refresh-1","token_type":"Bearer"}`))
	}))
	defer server.Close()

	state := "state-1"
	store := &memoryTokenStore{
		saveTokenErr: errors.New("token store failed"),
	}

	_, err := Service{
		Exchanger: Exchanger{
			HTTPClient: server.Client(),
			TokenURL:   server.URL,
		},
		Store:       store,
		In:          strings.NewReader("http://localhost:14573/callback?code=code-1&state=" + state + "\n"),
		Out:         &bytes.Buffer{},
		StateSource: func() string { return state },
	}.Login(context.Background(), LoginInput{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  "http://localhost:14573/callback",
	})
	if err == nil {
		t.Fatal("Login() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "token store failed") {
		t.Fatalf("error = %q, want token store failure", err)
	}
	if store.deleteSecretCalls != 1 {
		t.Fatalf("DeleteClientSecret() calls = %d, want 1", store.deleteSecretCalls)
	}
	if store.clientSecret != "" {
		t.Fatalf("client secret = %q, want cleared", store.clientSecret)
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

func TestServiceLoginRejectsMismatchedRedirectQuery(t *testing.T) {
	store := &memoryTokenStore{}

	_, err := Service{
		Store:       store,
		In:          strings.NewReader("http://localhost:14573/callback?tenant=wrong&code=code-1&state=state-1\n"),
		Out:         &bytes.Buffer{},
		StateSource: func() string { return "state-1" },
	}.Login(context.Background(), LoginInput{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  "http://localhost:14573/callback?tenant=expected",
	})
	if err == nil {
		t.Fatal("Login() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "redirect does not match") {
		t.Fatalf("error = %q, want redirect mismatch", err)
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
	got := BuildAuthorizeURL("https://ticktick.com/oauth/authorize", OAuthConfig{
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

func TestServiceLoginUsesConfiguredAuthorizeURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"access-1","refresh_token":"refresh-1","token_type":"Bearer"}`))
	}))
	defer server.Close()

	state := "state-1"
	callbackURL := "http://localhost:14573/callback?code=code-1&state=" + state
	store := &memoryTokenStore{}
	browser := &stubBrowser{}
	out := &bytes.Buffer{}

	_, err := Service{
		AuthorizeURL: "https://dida365.com/oauth/authorize",
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
	if len(browser.urls) != 1 {
		t.Fatalf("OpenURL() calls = %d, want 1", len(browser.urls))
	}
	if got, want := browser.urls[0], "https://dida365.com/oauth/authorize?client_id=client-1&redirect_uri=http%3A%2F%2Flocalhost%3A14573%2Fcallback&response_type=code&scope=tasks%3Aread+tasks%3Awrite&state=state-1"; got != want {
		t.Fatalf("OpenURL() = %q, want %q", got, want)
	}
}

func TestServiceLoginAcceptsAutomaticLocalCallback(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.PostForm.Get("code"); got != "code-1" {
			t.Fatalf("code = %q, want code-1", got)
		}
		_, _ = w.Write([]byte(`{"access_token":"access-1","refresh_token":"refresh-1","token_type":"Bearer"}`))
	}))
	defer tokenServer.Close()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	addr := l.Addr().String()
	_ = l.Close()

	state := "state-1"
	store := &memoryTokenStore{}
	browser := &stubBrowser{}
	out := &bytes.Buffer{}
	manualIn, manualWriter := io.Pipe()
	defer manualWriter.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for len(browser.urls) == 0 {
			time.Sleep(10 * time.Millisecond)
		}
		callbackURL := "http://" + addr + "/callback?code=code-1&state=" + state
		resp, err := http.Get(callbackURL)
		if err != nil {
			t.Errorf("http.Get() error = %v", err)
			return
		}
		_ = resp.Body.Close()
	}()

	token, err := Service{
		Exchanger: Exchanger{
			HTTPClient: tokenServer.Client(),
			TokenURL:   tokenServer.URL,
		},
		Store:       store,
		Browser:     browser,
		In:          manualIn,
		Out:         out,
		StateSource: func() string { return state },
	}.Login(context.Background(), LoginInput{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  "http://" + addr + "/callback",
	})
	manualWriter.Close()
	<-done
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if token.AccessToken != "access-1" {
		t.Fatalf("AccessToken = %q, want access-1", token.AccessToken)
	}
	if !strings.Contains(out.String(), "Waiting for browser callback") {
		t.Fatalf("output = %q, want automatic callback guidance", out.String())
	}
}

func TestServiceLoginFallsBackToManualWhenCallbackListenerCannotStart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"access-1","refresh_token":"refresh-1","token_type":"Bearer"}`))
	}))
	defer server.Close()

	state := "state-1"
	store := &memoryTokenStore{}
	out := &bytes.Buffer{}

	token, err := Service{
		Exchanger: Exchanger{
			HTTPClient: server.Client(),
			TokenURL:   server.URL,
		},
		Store:       store,
		In:          strings.NewReader("http://127.0.0.1:14573/callback?code=code-1&state=" + state + "\n"),
		Out:         out,
		StateSource: func() string { return state },
		Listen: func(network, address string) (net.Listener, error) {
			return nil, errors.New("bind failed")
		},
	}.Login(context.Background(), LoginInput{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  "http://127.0.0.1:14573/callback",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if token.AccessToken != "access-1" {
		t.Fatalf("AccessToken = %q, want access-1", token.AccessToken)
	}
	if !strings.Contains(out.String(), "Automatic callback unavailable") {
		t.Fatalf("output = %q, want fallback warning", out.String())
	}
	if !strings.Contains(out.String(), "Paste the full callback URL") {
		t.Fatalf("output = %q, want manual fallback prompt", out.String())
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

func TestServiceStatusIncludesTokenExpiry(t *testing.T) {
	now := time.Unix(1_776_351_966, 0).UTC()
	service := Service{
		Store: &memoryTokenStore{
			token: Token{
				AccessToken:   "access-1",
				RefreshToken:  "refresh-1",
				ExpiresIn:     3600,
				CreatedAtUnix: now.Unix(),
				ExpiresAtUnix: now.Add(time.Hour).Unix(),
			},
		},
		Now: func() time.Time { return now },
	}

	status, err := service.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if !status.ExpiryKnown {
		t.Fatal("ExpiryKnown = false, want true")
	}
	if status.ExpiresAtUnix != now.Add(time.Hour).Unix() {
		t.Fatalf("ExpiresAtUnix = %d, want %d", status.ExpiresAtUnix, now.Add(time.Hour).Unix())
	}
	if status.ExpiresInSeconds != 3600 {
		t.Fatalf("ExpiresInSeconds = %d, want 3600", status.ExpiresInSeconds)
	}
}

func TestServiceAccessTokenRefreshesExpiredToken(t *testing.T) {
	now := time.Unix(1_776_351_966, 0).UTC()
	store := &memoryTokenStore{
		token: Token{
			AccessToken:   "expired-access",
			RefreshToken:  "refresh-1",
			ExpiresIn:     3600,
			CreatedAtUnix: now.Add(-2 * time.Hour).Unix(),
			ExpiresAtUnix: now.Add(-time.Hour).Unix(),
		},
		clientSecret: "secret-1",
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.PostForm.Get("grant_type"); got != "refresh_token" {
			t.Fatalf("grant_type = %q, want refresh_token", got)
		}
		if got := r.PostForm.Get("refresh_token"); got != "refresh-1" {
			t.Fatalf("refresh_token = %q, want refresh-1", got)
		}
		_, _ = w.Write([]byte(`{"access_token":"fresh-access","token_type":"Bearer","expires_in":1800}`))
	}))
	defer server.Close()

	service := Service{
		ClientID: "client-1",
		Exchanger: Exchanger{
			HTTPClient: server.Client(),
			TokenURL:   server.URL,
			Now:        func() time.Time { return now },
		},
		Store: store,
		Now:   func() time.Time { return now },
	}

	accessToken, err := service.AccessToken(context.Background())
	if err != nil {
		t.Fatalf("AccessToken() error = %v", err)
	}
	if accessToken != "fresh-access" {
		t.Fatalf("AccessToken() = %q, want fresh-access", accessToken)
	}
	if store.saveTokenCalls != 1 {
		t.Fatalf("SaveToken() calls = %d, want 1", store.saveTokenCalls)
	}
	if store.token.RefreshToken != "refresh-1" {
		t.Fatalf("RefreshToken = %q, want original refresh-1 preserved", store.token.RefreshToken)
	}
	if store.token.ExpiresAtUnix != now.Add(30*time.Minute).Unix() {
		t.Fatalf("ExpiresAtUnix = %d, want %d", store.token.ExpiresAtUnix, now.Add(30*time.Minute).Unix())
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
