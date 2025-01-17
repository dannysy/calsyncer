package cmd

import (
	"calsyncer/internal/domain"
	"calsyncer/internal/exporter"
	"calsyncer/internal/importer"
	"context"
)

func noopCmd(ctx context.Context) {
	useCase := domain.New(ctx, &importer.Noop{}, &exporter.Noop{})
	useCase.TaskSync("1-59 * 8-20 * * 1-5 *")
	useCase.Stop()
}
