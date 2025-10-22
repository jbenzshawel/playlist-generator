package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Clients `json:"clients"`
}

type Clients struct {
	IowaPublicRadio Client `json:"ipr"`
}

type Client struct {
	BaseURL string `json:"baseURL"`
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
