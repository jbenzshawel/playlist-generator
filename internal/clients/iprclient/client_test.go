package iprclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetSongs(t *testing.T) {
	c := New(Config{})

	col, err := c.GetSongs(t.Context(), "2025-10-19")
	require.NoError(t, err)
	assert.Greater(t, len(col.Items), 0)
}
