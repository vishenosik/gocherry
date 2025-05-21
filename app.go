package app

import (
	"context"
	"log/slog"

	"github.com/vishenosik/gocherry/pkg/logs"

	webctx "github.com/vishenosik/web/context"
)

type App struct {
	log      *slog.Logger
	services []Service
}

type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type AppOption = func(*App)

func NewApp(opts ...AppOption) (*App, error) {

	log := logs.SetupLogger()

	app := &App{
		log: log,
	}

	for _, opt := range opts {
		opt(app)
	}

	return app, nil
}

func (app *App) Start(ctx context.Context) error {

	app.log.Info("start app")

	for _, service := range app.services {
		go service.Start(ctx)
	}
	return nil
}

func (app *App) AddServices(services ...Service) {
	if len(app.services) == 0 {
		app.services = make([]Service, 0, len(services))
	}
	app.services = append(app.services, services...)
}

func (app *App) Stop(ctx context.Context) {

	const msg = "app stopping"

	signal, ok := webctx.StopFromCtx(ctx)
	if ok {
		app.log.Info(msg, slog.String("signal", signal.Signal.String()))
	} else {
		app.log.Info(msg)
	}

	for _, service := range app.services {
		service.Stop(ctx)
	}

	app.log.Info("app stopped")
}
