package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jbenzshawel/playlist-generator/internal/common/cerror"
)

func TestNewHash(t *testing.T) {
	soWhatHash := "f38e7dcfa59e189ecb1a13eb7d96d30b7e8dd1c6a6d36c6c1f7a7fc83a59fa03"

	testCases := []struct {
		name         string
		artist       string
		track        string
		album        string
		expectedHash string
		expectedErr  error
	}{
		{
			name:         "valid inputs",
			artist:       "Miles Davis",
			track:        "So What",
			album:        "Kind of Blue",
			expectedHash: soWhatHash,
		},
		{
			name:         "ignores case",
			artist:       "MiLES Davis",
			track:        "SO WHAT",
			album:        "kind OF bLUE",
			expectedHash: soWhatHash,
		},
		{
			name:  "empty artist",
			track: "SO WHAT",
			album: "kind OF bLUE",
			expectedErr: cerror.NewValidationError("invalid song hash info", map[string]string{
				"artist": "argument cannot be empty",
			}),
		},
		{
			name:   "empty track",
			artist: "MiLES Davis",
			album:  "kind OF bLUE",
			expectedErr: cerror.NewValidationError("invalid song hash info", map[string]string{
				"track": "argument cannot be empty",
			}),
		},
		{
			name:   "empty album",
			artist: "MiLES Davis",
			track:  "kind OF bLUE",
			expectedErr: cerror.NewValidationError("invalid song hash info", map[string]string{
				"album": "argument cannot be empty",
			}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewSongHash(tc.artist, tc.track, tc.album)
			assert.Equal(t, tc.expectedHash, actual)
			assert.Equal(t, err, tc.expectedErr)
		})
	}
}
