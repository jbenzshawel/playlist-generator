package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSong(t *testing.T) {
	var (
		artist = "Spoon"
		track  = "Inside Out"
		album  = "They Want My Soul"
		upc    = "upc"
	)

	expectedHash, err := NewSongHash(artist, track, album)
	require.NoError(t, err)

	actual, err := NewSong(artist, track, album, upc)
	assert.NoError(t, err)
	assert.NotEmpty(t, actual.ID())
	assert.Equal(t, artist, actual.Artist())
	assert.Equal(t, track, actual.Track())
	assert.Equal(t, album, actual.Album())
	assert.Equal(t, upc, actual.UPC())
	assert.Equal(t, expectedHash, actual.SongHash())
	assert.NotEmpty(t, actual.Created())
}
