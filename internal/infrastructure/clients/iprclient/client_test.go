package iprclient

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetSongs(t *testing.T) {
	baseURL, err := url.Parse("https://api.composer.nprstations.org/v1/widget/51827818e1c8c2244542ab7b")
	require.NoError(t, err)
	c := New(Config{
		BaseURL: baseURL,
	})

	col, err := c.GetSongs(t.Context(), "2025-10-19")
	require.NoError(t, err)
	assert.Greater(t, len(col.Items), 0)
}
