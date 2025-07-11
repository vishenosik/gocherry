package gocherry

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/vishenosik/gocherry/pkg/config"
	"github.com/vishenosik/gocherry/pkg/logs"

	_ctx "github.com/vishenosik/gocherry/pkg/context"
)

var (
	BuildDate string
	GitBranch string
	GitCommit string
	GoVersion string
	GitTag    string
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

	if len(app.services) == 0 {
		app.services = make([]Service, 0, len(services))
	}

	if len(app.closers) == 0 {
		app.closers = make([]Closer, 0, len(services))
	}

	for _, service := range services {
		switch srv := service.(type) {
		case Service:
			app.services = append(app.services, srv)
		case Closer:
			app.closers = append(app.closers, srv)
		}
	}
}

func (app *App) Start(ctx context.Context) error {

	app.Log.Info("start app")

	for _, service := range app.services {
		go service.Start(ctx)
	}
	return nil
}

func (app *App) Stop(ctx context.Context) error {

	const msg = "app stopping"

	signal, ok := _ctx.StopFromCtx(ctx)
	if ok {
		app.Log.Info(msg, slog.String("signal", signal.Signal.String()))
	} else {
		app.Log.Info(msg)
	}

	for _, service := range app.services {
		service.Stop(ctx)
	}

	for _, closer := range app.closers {
		closer.Close(ctx)
	}

	app.Log.Info("app stopped")
	return nil
}

func ConfigFlags(structs ...any) {

	_structs := append(structs, config.Structs()...)

	flag.BoolFunc(
		"config.info",
		"Show config schema information",
		config.ConfigInfoEnv(os.Stdout, _structs...),
	)

	flag.Func(
		"config.gen",
		"Generate config schema",
		config.ConfigGenEnv(_structs...),
	)

}
