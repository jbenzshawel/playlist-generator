package sources

import (
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/sources/spinitron"
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/sources/studioone"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type studioOneClient interface {
	studioone.Client
}

type spinitronClient interface {
	spinitron.Client
}

type Commands struct {
	StudioOne studioone.Commands
	KRUI      spinitron.Commands
	KCCK      spinitron.Commands
}

func NewCommands(studioOne studioOneClient, spinClient spinitronClient, repository domain.Repository) Commands {
	return Commands{
		StudioOne: studioone.NewCommands(studioOne, repository),
		KRUI:      spinitron.NewCommands(spinClient, domain.KRUISourceType, repository),
		KCCK:      spinitron.NewCommands(spinClient, domain.KCCKSourceType, repository),
	}
}
