package playlists

import (
	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/playlists/services"
)

type Commands struct {
	CreatePlaylist       CreatePlaylistCommandHandler
	RandomTracksPlaylist RandomTracksPlaylistCommandHandler
	SearchTracks         SearchTracksCommandHandler
	SyncPlaylist         SyncPlaylistCommandHandler
}

func NewCommands(client services.Client, repository domain.Repository) Commands {
	playlistService := services.NewPlaylistService(client)
	searchService := services.NewSearchService(client)

	return Commands{
		CreatePlaylist:       NewCreatePlaylistCommand(playlistService, repository),
		RandomTracksPlaylist: NewRandomTracksPlaylistCommand(playlistService, repository),
		SearchTracks:         NewSearchTracksCommand(searchService, repository),
		SyncPlaylist:         NewSyncPlaylistCommand(playlistService, repository),
	}
}
