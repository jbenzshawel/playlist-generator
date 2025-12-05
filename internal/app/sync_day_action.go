package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/playlists"
	"github.com/jbenzshawel/playlist-generator/internal/sources"
)

type syncDayResult struct {
	SourceType   domain.SourceType
	PlaylistName string
	TotalTracks  int
	TracksAdded  int
}

func (a Application) syncDayAction(ctx context.Context, sourceType domain.SourceType, date string) (syncDayResult, error) {
	slog.Info("adding songs to Spotify playlist",
		slog.String("date", date),
		slog.String("source", sourceType.String()),
	)

	a.output.Section("Updating %s (%s) playlist for songs played on %s:", sourceType.String(), sourceType.Description(), date)

	lookupDone := a.output.InfoSpinner("Looking up songs...")

	listRes, err := a.Sources.ListSongs.Execute(ctx, sources.SourceSongListCommand{
		SourceType: sourceType,
		Date:       date,
	})
	if err != nil {
		return syncDayResult{}, fmt.Errorf("%s song list error: %w", sourceType, err)
	}

	lookupDone(fmt.Sprintf("%d songs found", listRes.FoundCount))

	progressBar := a.output.NewProgressBarCreator()

	searchRes, err := a.Playlists.SearchTracks.Execute(ctx, playlists.SearchTracksCommand{
		Progress: progressBar,
	})
	if err != nil {
		return syncDayResult{}, fmt.Errorf("spotify track update error: %w", err)
	}

	a.output.Info("%d matches found on Spotify (%d new songs searched)", searchRes.MatchedCount, searchRes.UnknownCount)

	createRes, err := a.Playlists.CreatePlaylist.Execute(ctx, playlists.CreatePlaylistCommand{
		Date:       date,
		SourceType: sourceType,
	})
	if err != nil {
		return syncDayResult{}, fmt.Errorf("create spotify playlist error: %w", err)
	}

	a.output.Info("%s playlist retrieved", createRes.Playlist.Name())

	syncDone := a.output.InfoSpinner("Adding new tracks to playlist...")

	syncRes, err := a.Playlists.SyncPlaylist.Execute(ctx, playlists.SyncPlaylistCommand{
		Playlist:   createRes.Playlist,
		SourceType: sourceType,
		Date:       date,
	})
	if err != nil {
		return syncDayResult{}, fmt.Errorf("sync spotify playlist error: %w", err)
	}

	syncDone(fmt.Sprintf("%d new tracks added", syncRes.NewTracks))

	a.output.Success("%s sync complete!", sourceType.String())

	return syncDayResult{
		SourceType:   sourceType,
		PlaylistName: createRes.Playlist.Name(),
		TotalTracks:  syncRes.TotalTracks,
		TracksAdded:  syncRes.NewTracks,
	}, nil
}
