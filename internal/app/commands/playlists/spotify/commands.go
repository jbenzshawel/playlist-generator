package spotify

import "github.com/jbenzshawel/playlist-generator/internal/domain"

type Client interface {
	trackSearcher
	creator
}

type Commands struct {
	UpdateTracks   UpdateTracksCommandHandler
	CreatePlaylist CreatePlaylistCommandHandler
}

func NewCommands(client Client, repository domain.Repository) Commands {
	return Commands{
		UpdateTracks:   NewUpdateTracksCommandHandler(client, repository),
		CreatePlaylist: NewCreatePlaylistCommand(client, repository),
	}
}
