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

	a.outputSection("Updating %s (%s) playlist for songs played on %s:", sourceType.String(), sourceType.Description(), date)

	listRes, err := a.Sources.ListSongs.Execute(ctx, sources.SourceSongListCommand{
		SourceType: sourceType,
		Date:       date,
	})
	if err != nil {
		return fmt.Errorf("%s song list error: %w", sourceType, err)
	}

	a.outputInfo("%d songs found", listRes.FoundCount)

	progressBar := a.outputCreateProgressBar()

	searchRes, err := a.Playlists.SearchTracks.Execute(ctx, playlists.SearchTracksCommand{
		Progress: progressBar,
	})
	if err != nil {
		return fmt.Errorf("spotify track update error: %w", err)
	}

	a.outputInfo("%d matches found on spotify (%d new songs searched)", searchRes.MatchedCount, searchRes.UnknownCount)

	createRes, err := a.Playlists.CreatePlaylist.Execute(ctx, playlists.CreatePlaylistCommand{
		Date:       date,
		SourceType: sourceType,
	})
	if err != nil {
		return fmt.Errorf("create spotify playlist error: %w", err)
	}

	a.outputInfo("%s playlist retrieved", createRes.Playlist.Name())

	syncRes, err := a.Playlists.SyncPlaylist.Execute(ctx, playlists.SyncPlaylistCommand{
		Playlist:   createRes.Playlist,
		SourceType: sourceType,
		Date:       date,
	})
	if err != nil {
		return fmt.Errorf("sync spotify playlist error: %w", err)
	}

	a.outputInfo("%d new tracks added", syncRes.NewTracks)

	a.outputSuccess("%s sync complete!", sourceType.String())

	return nil
}
