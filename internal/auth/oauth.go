package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type Exchanger struct {
	HTTPClient *http.Client
	TokenURL   string
}

func BuildAuthorizeURL(cfg OAuthConfig, state string) string {
	values := url.Values{}
	values.Set("client_id", cfg.ClientID)
	values.Set("scope", "tasks:read tasks:write")
	values.Set("state", state)
	values.Set("redirect_uri", cfg.RedirectURL)
	values.Set("response_type", "code")
	return "https://ticktick.com/oauth/authorize?" + values.Encode()
}

func (e Exchanger) ExchangeCode(ctx context.Context, cfg OAuthConfig, code string) (Token, error) {
	tokenURL := e.TokenURL
	if tokenURL == "" {
		tokenURL = "https://ticktick.com/oauth/token"
	}
	values := url.Values{}
	values.Set("code", code)
	values.Set("grant_type", "authorization_code")
	values.Set("scope", "tasks:read tasks:write")
	values.Set("redirect_uri", cfg.RedirectURL)

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

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return Token{}, err
	}
	return token, nil
}
