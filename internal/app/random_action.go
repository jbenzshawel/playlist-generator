package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jbenzshawel/playlist-generator/internal/playlists"
)

func (a Application) randomAction(ctx context.Context, numTracks int) error {
	slog.Info("updating random playlist with new randomAction tracks", slog.Int("numTracks", numTracks))

	done := a.output.Spinner(
		fmt.Sprintf("Updating Random Radio Playlist with %d new tracks", numTracks),
		fmt.Sprintf("Random Radio Playlist updated with %d new tracks!", numTracks),
	)

	_, err := a.Playlists.RandomTracksPlaylist.Execute(ctx, playlists.RandomTracksPlaylistCommand{
		NumTracks: numTracks,
	})
	if err != nil {
		return err
	}

	done()

	return nil
}
