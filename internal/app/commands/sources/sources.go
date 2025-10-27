package sources

import (
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/sources/studioone"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type client interface {
	studioone.Client
}

type Commands struct {
	StudioOne studioone.Commands
}

func NewCommands(client client, repository domain.Repository) Commands {
	return Commands{
		StudioOne: studioone.NewCommands(client, repository),
	}
}
