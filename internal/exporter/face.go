package exporter

import (
	"calsyncer/internal/importer"
)

type CalExporter interface {
	Set(calendar importer.Calendar) error
}
