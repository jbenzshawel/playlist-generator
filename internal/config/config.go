package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Clients `json:"clients"`
}

type Clients struct {
	IowaPublicRadio Client      `json:"ipr"`
	SpotifyClient   OAuthClient `json:"spotify"`
}

type Client struct {
	BaseURL string `json:"baseURL"`
}

type OAuthClient struct {
	Client
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	AuthURL      string `json:"authURL"`
	TokenURL     string `json:"tokenURL"`
}

func Load() (Config, error) {
	cfgBytes, err := os.ReadFile("config.json")
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
