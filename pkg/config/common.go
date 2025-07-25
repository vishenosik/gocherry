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
	return fmt.Sprintf("%s:%d", srv.Host, srv.Port)
}

type Credentials struct {
	User     string
	Password string
}

func ReadConfig(conf any) error {
	if err := cleanenv.ReadConfig(".env", conf); err != nil {
		return cleanenv.ReadEnv(conf)
	}
	return nil
}
