package exporter

import (
	"calsyncer/internal/importer"

	"github.com/rs/zerolog/log"
)

var _ CalExporter = (*Noop)(nil)

type Noop struct{}

func (e *Noop) Set(_ importer.Calendar) error {
	log.Info().Msg("noop exporter set events call")
	return nil
}
