package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type Exchanger struct {
	HTTPClient *http.Client
	TokenURL   string
	Now        func() time.Time
}

func BuildAuthorizeURL(authorizeURL string, cfg OAuthConfig, state string) string {
	values := url.Values{}
	values.Set("client_id", cfg.ClientID)
	values.Set("scope", "tasks:read tasks:write")
	values.Set("state", state)
	values.Set("redirect_uri", cfg.RedirectURL)
	values.Set("response_type", "code")
	if authorizeURL == "" {
		authorizeURL = "https://ticktick.com/oauth/authorize"
	}
	return authorizeURL + "?" + values.Encode()
}

func (e Exchanger) ExchangeCode(ctx context.Context, cfg OAuthConfig, code string) (Token, error) {
	values := url.Values{}
	values.Set("code", code)
	values.Set("grant_type", "authorization_code")
	values.Set("scope", "tasks:read tasks:write")
	values.Set("redirect_uri", cfg.RedirectURL)
	return e.exchange(ctx, cfg, values)
}

func (e Exchanger) RefreshToken(ctx context.Context, cfg OAuthConfig, refreshToken string) (Token, error) {
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", refreshToken)
	return e.exchange(ctx, cfg, values)
}

func (e Exchanger) exchange(ctx context.Context, cfg OAuthConfig, values url.Values) (Token, error) {
	tokenURL := e.TokenURL
	if tokenURL == "" {
		tokenURL = "https://ticktick.com/oauth/token"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.SetBasicAuth(cfg.ClientID, cfg.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := e.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Token{}, fmt.Errorf("oauth token exchange failed: %s", resp.Status)
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return Token{}, err
	}
	if token.AccessToken == "" {
		return Token{}, errors.New("oauth token response missing access_token")
	}
	return token.withExpiry(e.now()), nil
}

func (e Exchanger) now() time.Time {
	if e.Now != nil {
		return e.Now()
	}
	return time.Now().UTC()
}
