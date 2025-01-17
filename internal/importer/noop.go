package importer

import (
	"github.com/rs/zerolog/log"
)

var _ CalImporter = (*Noop)(nil)

type Noop struct {
}

func (i *Noop) Get() (Calendar, error) {
	log.Info().Msg("noop importer get events call")
	return Calendar{}, nil
}
