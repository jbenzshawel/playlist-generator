package studioone

import (
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/sources/internal"
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/sources/studioone/internal/providers"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type Client interface {
	providers.SongGetter
}

func NewListSongsCommand(client Client, repository domain.Repository) internal.SongListCommandHandler {
	return internal.NewSongListCommand(providers.NewSongProvider(client), repository)
}
