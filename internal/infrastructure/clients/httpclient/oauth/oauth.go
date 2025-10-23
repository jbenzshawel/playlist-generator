package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

type AuthenticatorConfig struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	RedirectURL  string
	Scopes       []string
}

func NewAuthenticator(cfg AuthenticatorConfig) *authenticator {
	return &authenticator{
		oauthCfg: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  cfg.AuthURL,
				TokenURL: cfg.TokenURL,
			},
			RedirectURL: cfg.RedirectURL,
			Scopes:      cfg.Scopes,
		},
	}
}

func (a *authenticator) AuthCodeURL() (string, error) {
	var err error
	a.state, err = generateState()
	if err != nil {
		return "", err
	}
	return a.oauthCfg.AuthCodeURL(a.state), nil
}

type authenticator struct {
	oauthCfg *oauth2.Config
	state    string
}

func (a *authenticator) GetAuthCodeCallbackHandler(chOAuthClient chan *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok, err := a.token(r.Context(), a.state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			return
		}
		if st := r.FormValue("state"); st != a.state {
			http.NotFound(w, r)
			return
		}

		client := a.oauthCfg.Client(r.Context(), tok)
		chOAuthClient <- client
	}
}

func (a *authenticator) token(ctx context.Context, state string, r *http.Request) (*oauth2.Token, error) {
	values := r.URL.Query()
	if e := values.Get("error"); e != "" {
		return nil, fmt.Errorf("spotify auth failed: %v", e)
	}
	code := values.Get("code")
	if code == "" {
		return nil, errors.New("spotify access code empty")
	}
	actualState := values.Get("state")
	if actualState != state {
		return nil, errors.New("spotify redirect state mismatch")
	}
	return a.oauthCfg.Exchange(ctx, code)
}

func generateState() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %v", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
