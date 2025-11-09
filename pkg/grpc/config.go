package grpc

import (
	"time"

	"github.com/vishenosik/gocherry/pkg/config"
)

func init() {
	config.AddStructs(ConfigEnv{})
}

type ConfigEnv struct {
	Port    uint16        `env:"GRPC_PORT" env-default:"9090" desc:"grpc server port"`
	Timeout time.Duration `env:"GRPC_TIMEOUT" env-default:"15s" desc:"grpc timeout"`
}

func (ConfigEnv) Desc() string {
	return "gRPC server settings"
}
