package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/jbenzshawel/playlist-generator/internal/common/cerror"
)

func newSongHash(artist, track, album string) (string, error) {
	if err := validateSongHash(artist, track, album); err != nil {
		return "", err
	}

	song := fmt.Sprintf("%s-%s-%s", artist, track, album)
	hashBytes := sha256.Sum256([]byte(strings.ToLower(song)))

	return hex.EncodeToString(hashBytes[:]), nil
}

func validateSongHash(artist string, track string, album string) error {
	fieldErrors := make(map[string]string)
	if len(artist) == 0 {
		fieldErrors["artist"] = "argument cannot be empty"
	}
	if len(track) == 0 {
		fieldErrors["track"] = "argument cannot be empty"
	}
	if len(album) == 0 {
		fieldErrors["album"] = "argument cannot be empty"
	}

	if len(fieldErrors) > 0 {
		return cerror.NewValidationError("invalid song hash info", fieldErrors)
	}
	return nil
}
