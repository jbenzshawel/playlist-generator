package spotify

import (
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/internal/providers"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type Client interface {
	providers.TrackSearcher
	creator
	playlist
}

type Commands struct {
	SearchTracks   SearchTracksCommandHandler
	CreatePlaylist CreatePlaylistCommandHandler
	SyncPlaylist   SyncPlaylistCommandHandler
}

func NewCommands(client Client, repository domain.Repository) Commands {
	return Commands{
		SearchTracks:   NewSearchTracksCommand(client, repository),
		CreatePlaylist: NewCreatePlaylistCommand(client, repository),
		SyncPlaylist:   NewSyncPlaylistCommand(client, repository),
	}
}
