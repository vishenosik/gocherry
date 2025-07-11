package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/vishenosik/gocherry/pkg/config"
	"github.com/vishenosik/gocherry/pkg/logs"
)

func appComponent() slog.Attr {
	return logs.AppComponent("http")
}

type Server struct {
	log    *slog.Logger
	server *http.Server
	config Config
}

func init() {
	config.AddStructs(ConfigEnv{})
}

type ConfigEnv struct {
	Port    uint16        `env:"HTTP_PORT" default:"8080" desc:"HTTP server port"`
	Timeout time.Duration `env:"HTTP_TIMEOUT" default:"15s" desc:"HTTP timeout"`
}

func (ConfigEnv) Desc() string {
	return "http server settings"
}

type Config struct {
	Server config.Server
}

type ServerOption func(*Server)

func NewHttpServer(
	handler http.Handler,
	opts ...ServerOption,
) (*Server, error) {

	if handler == nil {
		return nil, errors.New("handler can't be nil")
	}

	log := logs.SetupLogger().With(appComponent())

	var envConf ConfigEnv
	if err := config.ReadConfig(&envConf); err != nil {
		log.Warn("init http server: failed to read config", logs.Error(err))
	}

	config := Config{
		Server: config.Server{
			Port:    envConf.Port,
			Timeout: envConf.Timeout,
		},
	}

	srv := &Server{
		log: log,
		server: &http.Server{
			Addr:    config.Server.String(),
			Handler: handler,
		},
		config: config,
	}

	for _, opt := range opts {
		opt(srv)
	}

	if err := validateConfig(config); err != nil {
		return nil, errors.Wrap(err, "failed to validate http app config")
	}

	return srv, nil
}

func (a *Server) Start(_ context.Context) error {
	const op = "http.Server.Start"

	log := a.log.With(logs.Operation(op), slog.Any("port", a.config.Server.Port))

	log.Info("starting server")

	if err := a.server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return errors.Wrap(err, op)
		}
	}

	return nil
}

func (a *Server) Stop(ctx context.Context) error {

	const op = "http.Server.Stop"

	a.log.Info("stopping server", logs.Operation(op), slog.Any("port", a.config.Server.Port))

	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error("server shutdown failed", logs.Error(err))
	}
	return nil
}

func validateConfig(config Config) error {
	const op = "validateConfig"
	if err := config.Server.Validate(); err != nil {
		return errors.Wrap(err, op)
	}
	return nil
}
