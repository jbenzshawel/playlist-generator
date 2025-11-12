package spotify

import (
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/internal/services"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type Client interface {
	services.Client
}

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
