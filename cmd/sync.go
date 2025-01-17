package cmd

import (
	"calsyncer/internal/domain"
	"calsyncer/internal/exporter"
	"calsyncer/internal/importer"
	"context"
)

func syncCmd(ctx context.Context) {
	useCase := domain.New(ctx, importer.NewCalDAV(), exporter.NewFileExporter(ctx, exporter.NewTodoist()))
	useCase.TaskSync("0 */5 8-20 * * 1-5 *")
	defer useCase.Stop()
}
