package gocherry

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/vishenosik/gocherry/pkg/errors"
	"github.com/vishenosik/gocherry/pkg/logs"

	_ctx "github.com/vishenosik/gocherry/pkg/context"
)

type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Closer interface {
	Close(ctx context.Context) error
}

type App struct {
	Log      *slog.Logger
	services []Service
	closers  []Closer

	once sync.Once
}

type AppOption = func(*App)

func NewApp(opts ...AppOption) (*App, error) {

	log := logs.SetupLogger()

	app := &App{
		Log: log,
	}

	for _, opt := range opts {
		opt(app)
	}

	return app, nil
}

func (app *App) AddServices(services ...any) {

	app.once.Do(func() {
		app.services = make([]Service, 0, len(services))
		app.closers = make([]Closer, 0, len(services))
	})

	rejected := make([]string, 0)

	for _, service := range services {
		switch srv := service.(type) {
		case Service:
			app.services = append(app.services, srv)
		case Closer:
			app.closers = append(app.closers, srv)
		default:
			rejected = append(rejected, fmt.Sprintf("%T", srv))
		}
	}

	if len(rejected) > 0 {
		app.Log.Error("services with types doesn't implement gocherry.Service or gocherry.Closer interfaces",
			slog.Any("types", rejected),
		)
	}
}

func (app *App) Start(ctx context.Context) error {

	app.Log.Info("start app")

	for _, service := range app.services {
		go service.Start(ctx)
	}
	return nil
}

func (app *App) Stop(ctx context.Context) {

	const msg = "app stopping"

	result := new(errors.MultiError)

	signal, ok := _ctx.StopFromCtx(ctx)
	if ok {
		app.Log.Info(msg, slog.String("signal", signal.Signal.String()))
	} else {
		app.Log.Info(msg)
	}

	for _, service := range app.services {
		err := service.Stop(ctx)
		if err != nil {
			result.Append(err)
		}
	}

	for _, closer := range app.closers {
		err := closer.Close(ctx)
		if err != nil {
			result.Append(err)
		}
	}

	err := result.ErrorOrNil()
	if err != nil {
		app.Log.Error("app stopped with errors", logs.Error(err))
	} else {
		app.Log.Info("app stopped")
	}
}
