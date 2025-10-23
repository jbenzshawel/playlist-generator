package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient/decode"
)

type Config struct {
	AuthURL      *url.URL
	ClientID     string
	ClientSecret string
}

type token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type requestToken struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type TokenGetter struct {
	Cfg Config

	lock  sync.Mutex
	token *token
	ttl   time.Time
}

type client interface {
	Do(req *http.Request) (*http.Response, error)
}

func (a *TokenGetter) GetToken(ctx context.Context, client client) (string, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.token != nil && a.ttl.Before(time.Now()) {
		return a.token.AccessToken, nil
	}

	body := requestToken{
		GrantType:    "client_credentials",
		ClientID:     a.Cfg.ClientID,
		ClientSecret: a.Cfg.ClientSecret,
	}

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(body)
	if err != nil {
		return "", fmt.Errorf("failed to encode auth token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.Cfg.AuthURL.String(), &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create auth token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth token request failed: %w", err)
	}

	t, err := decode.JSON[token](resp)
	if err != nil {
		return "", fmt.Errorf("failed to decode auth token response: %w", err)
	}

	a.token = &t
	a.ttl = time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)

	return a.token.AccessToken, nil
}
