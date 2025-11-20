//go:build !prod

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func LoadForTest(t *testing.T) Config {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	configPath := ""
	for !strings.HasSuffix(dir, "playlist-generator") {
		configPath = filepath.Join(configPath, "../")
		dir = dir[0:strings.LastIndex(dir, "/")]
	}
	configPath = filepath.Join(configPath, "config.json")

	cfgBytes, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var cfg Config
	err = json.Unmarshal(cfgBytes, &cfg)
	require.NoError(t, err)

	return cfg
}
