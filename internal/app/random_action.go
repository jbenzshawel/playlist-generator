package app

import (
	"context"
	"log/slog"

	"github.com/jbenzshawel/playlist-generator/internal/playlists"
)

func (a Application) randomAction(ctx context.Context, numTracks int) error {
	slog.Info("updating random playlist with new randomAction tracks", slog.Int("numTracks", numTracks))

	_, err := a.Playlists.RandomTracksPlaylist.Execute(ctx, playlists.RandomTracksPlaylistCommand{
		NumTracks: numTracks,
	})
	if err != nil {
		return err
	}

	return nil
}
