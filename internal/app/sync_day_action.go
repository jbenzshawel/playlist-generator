package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/playlists"
	"github.com/jbenzshawel/playlist-generator/internal/sources"
)

func (a Application) syncDayAction(ctx context.Context, sourceType domain.SourceType, date string) error {
	slog.Info("adding songs to Spotify playlist",
		slog.String("date", date),
		slog.String("source", sourceType.String()),
	)

	_, err := a.Sources.ListSongs.Execute(ctx, sources.SourceSongListCommand{
		SourceType: sourceType,
		Date:       date,
	})
	if err != nil {
		return fmt.Errorf("%s song list error: %w", sourceType, err)
	}

	_, err = a.Playlists.SearchTracks.Execute(ctx, playlists.SearchTracksCommand{})
	if err != nil {
		return fmt.Errorf("spotify track update error: %w", err)
	}

	createRes, err := a.Playlists.CreatePlaylist.Execute(ctx, playlists.CreatePlaylistCommand{
		Date:       date,
		SourceType: sourceType,
	})
	if err != nil {
		return fmt.Errorf("create spotify playlist error: %w", err)
	}

	_, err = a.Playlists.SyncPlaylist.Execute(ctx, playlists.SyncPlaylistCommand{
		Playlist:   createRes.Playlist,
		SourceType: sourceType,
		Date:       date,
	})
	if err != nil {
		return fmt.Errorf("sync spotify playlist error: %w", err)
	}

	return err
}
