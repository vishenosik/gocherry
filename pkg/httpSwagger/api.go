package httpSwagger

import (
	"fmt"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/swaggo/swag/v2"
	"github.com/vishenosik/gocherry/pkg/config"
	"github.com/vishenosik/gocherry/pkg/logs"
)

func init() {
	config.AddStructs(&ConfigEnv{})
}

type ConfigEnv struct {
	// HTTP server port
	Port uint16 `env:"HTTP_PORT" env-default:"8080" desc:"-"`
	// HTTP host
	Host string `env:"SWAGGER_HTTP_HOST" env-default:"localhost:8080" desc:"HTTP host"`
	// Enable swagger
	Enable bool `env:"SWAGGER_ENABLE" env-default:"false" desc:"Enable swagger"`
}

func (ConfigEnv) Desc() string {
	return "Swagger settings"
}

type Swagger struct {
	enable bool
}

func NewSwagger(spec *swag.Spec) *Swagger {

	var envConf ConfigEnv
	if err := config.ReadConfig(&envConf); err != nil {
		log.Println("init http server: failed to read config", logs.Error(err))
	}

	s := &Swagger{
		enable: envConf.Enable,
	}

	spec.Schemes = []string{"http", "https"}
	spec.Host = fmt.Sprintf("localhost:%d", envConf.Port)

	if envConf.Host != "" {
		spec.Host = envConf.Host
	}

	return s
}

func (s *Swagger) Routers(r chi.Router) {
	if !s.enable {
		return
	}
	r.Group(func(r chi.Router) {
		r.Get("/swagger/*", Handler(
			URL("doc.json"),
		))
	})
}
