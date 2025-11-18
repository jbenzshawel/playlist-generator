package studioone

import (
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/sources/list"
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/sources/studioone/internal/providers"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type Commands struct {
	ListSongs list.SongListCommandHandler
}

type Client interface {
	providers.SongGetter
}

func NewCommands(client Client, repository domain.Repository) Commands {
	return Commands{
		ListSongs: list.NewSongListCommand(providers.NewSongProvider(client), repository),
	}
}
