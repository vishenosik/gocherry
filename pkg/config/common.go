package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Server struct {
	Host    string
	Port    uint16 `validate:"gte=1,lte=65535"`
	Timeout time.Duration
}

func (srv Server) Validate() error {
	if err := validator.New().Struct(srv); err != nil {
		return err
	}
	return nil
}

func (srv Server) String() string {
	if srv.Host == "" {
		srv.Host = "localhost"
	}
	return fmt.Sprintf("%s:%d", srv.Host, srv.Port)
}

type Credentials struct {
	User     string
	Password string
}

func ReadConfig(conf any) error {
	return cleanenv.ReadConfig(".env", conf)
}
