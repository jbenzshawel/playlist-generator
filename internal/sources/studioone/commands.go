package studioone

import (
	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/sources/internal"
	"github.com/jbenzshawel/playlist-generator/internal/sources/studioone/internal/providers"
)

type Client interface {
	providers.SongGetter
}

func NewListSongsCommand(client Client, repository domain.Repository) internal.SongListCommandHandler {
	return internal.NewSongListCommand(providers.NewSongProvider(client), repository)
}
