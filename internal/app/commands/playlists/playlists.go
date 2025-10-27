package playlists

import (
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type client interface {
	spotify.Client
}

type Commands struct {
	Spotify spotify.Commands
}

func NewCommands(client client, repository domain.Repository) Commands {
	return Commands{
		Spotify: spotify.NewCommands(client, repository),
	}
}
