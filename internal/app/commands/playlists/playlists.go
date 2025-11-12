package playlists

import (
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type Commands struct {
	Spotify spotify.Commands
}

func NewCommands(client spotify.Client, repository domain.Repository) Commands {
	return Commands{
		Spotify: spotify.NewCommands(client, repository),
	}
}
