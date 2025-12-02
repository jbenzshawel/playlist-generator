package sources

import (
	"context"
	"fmt"

	"github.com/jbenzshawel/playlist-generator/internal/common/decorator"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/sources/internal"
	"github.com/jbenzshawel/playlist-generator/internal/sources/spinitron"
	"github.com/jbenzshawel/playlist-generator/internal/sources/studioone"
)

type studioOneClient interface {
	studioone.Client
}

type spinitronClient interface {
	spinitron.Client
}

type Commands struct {
	ListSongs SongListCommandHandler
}

func NewCommands(studioOne studioOneClient, spinClient spinitronClient, repository domain.Repository) Commands {
	return Commands{
		ListSongs: &songListCommand{
			commands: map[domain.SourceType]internal.SongListCommandHandler{
				domain.StudioOneSourceType:        studioone.NewListSongsCommand(studioOne, repository),
				domain.KRUISourceType:             spinitron.NewListSongsCommand(spinClient, domain.KRUISourceType, repository),
				domain.KCCKSourceType:             spinitron.NewListSongsCommand(spinClient, domain.KCCKSourceType, repository),
				domain.KBEMSourceType:             spinitron.NewListSongsCommand(spinClient, domain.KBEMSourceType, repository),
				domain.KCSMSourceType:             spinitron.NewListSongsCommand(spinClient, domain.KCSMSourceType, repository),
				domain.EastVillageRadioSourceType: spinitron.NewListSongsCommand(spinClient, domain.EastVillageRadioSourceType, repository),
				domain.WKCRSourceType:             spinitron.NewListSongsCommand(spinClient, domain.WKCRSourceType, repository),
				domain.WDCBSourceType:             spinitron.NewListSongsCommand(spinClient, domain.WDCBSourceType, repository),
				domain.KUVOSourceType:             spinitron.NewListSongsCommand(spinClient, domain.KUVOSourceType, repository),
				domain.WSUMSourceType:             spinitron.NewListSongsCommand(spinClient, domain.WSUMSourceType, repository),
				domain.KZSCSourceType:             spinitron.NewListSongsCommand(spinClient, domain.KZSCSourceType, repository),
				domain.KSPCSourceType:             spinitron.NewListSongsCommand(spinClient, domain.KSPCSourceType, repository),
			},
		},
	}
}

type SourceSongListCommand struct {
	SourceType domain.SourceType
	Date       string
}

type SongListCommandHandler decorator.CommandHandler[SourceSongListCommand]

type songListCommand struct {
	commands map[domain.SourceType]internal.SongListCommandHandler
}

func (c *songListCommand) Execute(ctx context.Context, cmd SourceSongListCommand) (any, error) {
	command, ok := c.commands[cmd.SourceType]
	if !ok {
		return nil, fmt.Errorf("source type %s not supported", cmd.SourceType)
	}

	return command.Execute(ctx, internal.SongListCommand{
		Date: cmd.Date,
	})
}
