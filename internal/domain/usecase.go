package domain

import (
	"calsyncer/internal/exporter"
	"calsyncer/internal/importer"
	"context"

	"github.com/adhocore/gronx/pkg/tasker"
	"github.com/sourcegraph/conc/pool"
)

type UseCase struct {
	calImporter importer.CalImporter
	calExporter exporter.CalExporter
	pool        *pool.ContextPool
	ctx         context.Context
}

func New(ctx context.Context, callImporter importer.CalImporter, calExporter exporter.CalExporter) *UseCase {
	return &UseCase{
		calImporter: callImporter,
		calExporter: calExporter,
		pool:        pool.New().WithContext(ctx).WithMaxGoroutines(10),
		ctx:         ctx,
	}
}

func (uc *UseCase) SyncOnce() error {
	calendar, err := uc.calImporter.Get()
	if err != nil {
		return err
	}

	return uc.calExporter.Set(calendar)
}

func (uc *UseCase) TaskSync(cronExpr string) {
	taskr := tasker.New(tasker.Option{})
	taskr.Task(cronExpr, func(_ context.Context) (int, error) {
		return 0, uc.SyncOnce()
	})
	uc.pool.Go(func(ctx context.Context) error {
		taskr.Run()
		return nil
	})
}

func (uc *UseCase) Stop() {
	uc.pool.Wait()
}
