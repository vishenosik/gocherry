package gocherry

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/vishenosik/gocherry/pkg/config"
	"github.com/vishenosik/gocherry/pkg/http"
	"github.com/vishenosik/gocherry/pkg/logs"

	webctx "github.com/vishenosik/web/context"
)

type App struct {
	Log      *slog.Logger
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
		Log: log,
	}

	for _, opt := range opts {
		opt(app)
	}

	return app, nil
}

func (app *App) Start(ctx context.Context) error {

	app.Log.Info("start app")

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

func (app *App) Stop(ctx context.Context) error {

	const msg = "app stopping"

	signal, ok := webctx.StopFromCtx(ctx)
	if ok {
		app.Log.Info(msg, slog.String("signal", signal.Signal.String()))
	} else {
		app.Log.Info(msg)
	}

	for _, service := range app.services {
		service.Stop(ctx)
	}

	app.Log.Info("app stopped")
	return nil
}

func Flags() {

	structs := []any{
		logs.EnvConfig{},
		http.EnvConfig{},
	}

	flag.BoolFunc(
		"config.info",
		"Show config schema information",
		config.ConfigInfo(os.Stdout, structs...),
	)

	flag.Func(
		"config.doc",
		"Update config example in docs",
		config.ConfigDoc(structs...),
	)

}
