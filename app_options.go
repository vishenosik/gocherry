package app

import (
	"net/http"

	_http "github.com/vishenosik/gocherry/pkg/http"
	"github.com/vishenosik/gocherry/pkg/logs"
)

func WithHttpRoutes(handler http.Handler) AppOption {
	return func(app *App) {
		if handler == nil {
			app.log.Warn("failed to add http service: handler is nil")
			return
		}
		app.AddServices(_http.NewHttpServer(
			app.log.With(logs.AppComponent("http")),
			handler,
		))
	}
}

func WithWorkerPool(subscriptions ...chan PoolTask) AppOption {
	return func(app *App) {
		pool, err := NewPool(
			app.log.With(logs.AppComponent("worker pool")),
			subscriptions...,
		)
		if err != nil {
			app.log.Warn("failed to init worker pool", logs.Error(err))
			return
		}
		app.AddServices(pool)
	}
}
