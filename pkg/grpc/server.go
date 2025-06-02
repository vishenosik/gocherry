package grpc

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/vishenosik/gocherry/pkg/config"
	"github.com/vishenosik/gocherry/pkg/logs"
)

func appComponent() slog.Attr {
	return logs.AppComponent("gRPC")
}

type Server struct {
	// log is a structured logger for the application.
	log *slog.Logger
	// server is the main gRPC server instance.
	server *grpc.Server
	config Config
}

type ConfigEnv struct {
	Port    uint16        `env:"GRPC_PORT" default:"9090" desc:"grpc server port"`
	Timeout time.Duration `env:"GRPC_TIMEOUT" default:"15s" desc:"grpc timeout"`
}

type Config struct {
	Server config.Server
}

type ServerOption func(*Server)

type GrpcService interface {
	RegisterService(server *grpc.Server)
}

func NewGrpcServer(
	services []GrpcService,
	opts ...ServerOption,
) (*Server, error) {

	log := logs.SetupLogger().With(appComponent())

	var envConf ConfigEnv
	if err := config.ReadConfig(&envConf); err != nil {
		log.Warn("init http server: failed to read config", logs.Error(err))
	}

	server := grpc.NewServer()

	for _, service := range services {
		service.RegisterService(server)
	}

	config := Config{
		Server: config.Server{
			Port:    envConf.Port,
			Timeout: envConf.Timeout,
		},
	}

	srv := &Server{
		log:    log,
		server: server,
		config: config,
	}

	for _, opt := range opts {
		opt(srv)
	}

	if err := validateConfig(config); err != nil {
		return nil, errors.Wrap(err, "failed to validate gRPC app config")
	}

	return srv, nil
}

func (a *Server) Start(_ context.Context) error {
	const op = "grpc.Server.Start"

	log := a.log.With(
		slog.String("op", op),
		slog.Any("port", a.config.Server.Port),
	)

	log.Info("starting gRPC server")

	listener, err := net.Listen("tcp", a.config.Server.String())
	if err != nil {
		return errors.Wrap(err, op)
	}

	log.Info("gRPC server is running", slog.String("addr", listener.Addr().String()))

	if err := a.server.Serve(listener); err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (a *Server) Stop(ctx context.Context) error {

	const op = "grpc.Server.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Any("port", a.config.Server.Port))

	a.server.GracefulStop()
	return nil
}

func validateConfig(config Config) error {
	const op = "validateConfig"
	if err := config.Server.Validate(); err != nil {
		return errors.Wrap(err, op)
	}
	return nil
}
