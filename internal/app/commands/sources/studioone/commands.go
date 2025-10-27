package studioone

import "github.com/jbenzshawel/playlist-generator/internal/domain"

type Commands struct {
	ListSongs SongListCommandHandler
}

type Client interface {
	songGetter
}

func NewCommands(client Client, repository domain.Repository) Commands {
	return Commands{
		ListSongs: NewSongListCommand(client, repository),
	}
}
