package gocherry

import (
	"net/http"

	_http "github.com/vishenosik/gocherry/pkg/http"
	"github.com/vishenosik/gocherry/pkg/logs"
)

func WithHttpRoutes(handler http.Handler) AppOption {
	return func(app *App) {
		if handler == nil {
			app.Log.Warn("failed to add http service: handler is nil")
			return
		}

		server, err := _http.NewHttpServer(
			handler,
		)

		if err != nil {
			app.Log.Warn("failed to add http service", logs.Error(err))
			return
		}

		app.AddServices(server)
	}
}

func WithWorkerPool(subscriptions ...chan PoolTask) AppOption {
	return func(app *App) {
		pool, err := NewPool(
			subscriptions...,
		)
		if err != nil {
			app.Log.Warn("failed to init worker pool", logs.Error(err))
			return
		}
		app.AddServices(pool)
	}
}
