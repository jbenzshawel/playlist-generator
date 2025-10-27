package spotify

import "github.com/jbenzshawel/playlist-generator/internal/domain"

type Client interface {
	TrackSearcher
}

type Commands struct {
	UpdateTracks UpdateTracksCommandHandler
}

func NewCommands(client Client, repository domain.Repository) Commands {
	return Commands{
		UpdateTracks: NewUpdateTracksCommandHandler(client, repository),
	}
}
