package gocherry

import (
	"context"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/vishenosik/concurrency"
	"github.com/vishenosik/gocherry/pkg/logs"
)

type PoolTask struct {
	ID       string
	Func     func()
	Priority int
}

type Pool struct {
	log     *slog.Logger
	pool    *concurrency.Pool
	subChan <-chan PoolTask
}

func NewPool(subscriptions ...chan PoolTask) (*Pool, error) {
	return NewPoolContext(context.Background(), subscriptions...)
}

func NewPoolContext(ctx context.Context, subscriptions ...chan PoolTask) (*Pool, error) {
	if len(subscriptions) == 0 {
		return nil, errors.New("no subscriptions provided")
	}

	log := logs.SetupLogger().With(logs.AppComponent("worker pool"))

	pool := &Pool{
		log:     log,
		pool:    concurrency.NewWorkerPoolContext(ctx, concurrency.WithWorkersControl(3, 256, 3)),
		subChan: concurrency.MergeChannels(ctx, uint16(1024), subscriptions...),
	}

	return pool, nil
}

func (p *Pool) Start(ctx context.Context) error {
	p.pool.Start(ctx)

	metrics := p.pool.GetMetrics()

	p.log.Info("pool started",
		slog.Int("workers_current", int(metrics.WorkersCurrent)),
		slog.Int("workers_max", int(metrics.WorkersMax)),
		slog.Int("workers_min", int(metrics.WorkersMin)),
	)

	go func() {
		for task := range p.subChan {
			// p.log.Info("adding task",
			// 	slog.String("id", task.ID),
			// )

			_, err := p.pool.AddTask(
				concurrency.Task{
					ID:       task.ID,
					Func:     task.Func,
					Priority: concurrency.Priority(task.Priority),
				},
			)
			// p.log.Info("added task", slog.String("id", task.ID))
			if err != nil {
				p.log.Error("pool error", logs.Error(err))
				if errors.Is(err, concurrency.ErrPoolClosed) {
					// TODO: extra handling
					// TODO: log error?
					return
				}
			}
		}
		p.log.Warn("subs exited")
	}()
	return nil
}

func (p *Pool) Stop(ctx context.Context) error {
	p.pool.Stop(ctx)
	p.log.Info("pool stopped")
	return nil
}
