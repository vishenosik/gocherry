package http

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/vishenosik/gocherry/pkg/config"
	"github.com/vishenosik/gocherry/pkg/logs"
)

type Server struct {
	log    *slog.Logger
	server *http.Server
	port   uint16
}

type EnvConfig struct {
	Port uint16 `env:"REST_PORT" default:"8080" desc:"REST server port"`
}

type Config struct {
	Port    uint16 `validate:"gte=1,lte=65535"`
	Timeout time.Duration
}

func validateConfig(conf Config) error {
	const op = "validateConfig"
	valid := validator.New()
	if err := valid.Struct(conf); err != nil {
		return errors.Wrap(err, op)
	}
	return nil
}

func NewHttpServer(logger *slog.Logger, handler http.Handler) *Server {
	var envConf EnvConfig
	if err := config.ReadConfig(&envConf); err != nil {
		log.Println(errors.Wrap(err, "init http app: failed to read config"))
	}

	config := Config{
		Port: envConf.Port,
	}

	return NewHttpAppConfig(config, logger, handler)
}

func NewHttpAppConfig(config Config, logger *slog.Logger, handler http.Handler) *Server {

	if err := validateConfig(config); err != nil {
		panic(errors.Wrap(err, "failed to validate http app config"))
	}

	if logger == nil {
		panic("logger can't be nil")
	}

	if handler == nil {
		panic("handler can't be nil")
	}

	return &Server{
		log: logger,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", config.Port),
			Handler: handler,
		},
		port: config.Port,
	}
}

func (a *Server) Start(_ context.Context) error {
	const op = "http.Server.Run"

	log := a.log.With(logs.Operation(op), slog.Any("port", a.port))

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

	a.log.Info("stopping server", logs.Operation(op), slog.Any("port", a.port))

	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error("server shutdown failed", logs.Error(err))
	}
	return nil
}
