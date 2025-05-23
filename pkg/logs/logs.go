package logs

import (
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/vishenosik/gocherry/pkg/config"
	"github.com/vishenosik/web/colors"
)

const (
	EnvDev  = "dev"
	EnvProd = "prod"
	EnvTest = "test"
)

type EnvConfig struct {
	Env string `env:"ENV" default:"dev" desc:"The environment in which the application is running"`
}

type Config struct {
	Env        string `validate:"oneof=dev prod test"`
	Marshaller string `validate:"oneof=json yaml"`
}

func validateConfig(conf Config) error {
	const op = "validateConfig"
	valid := validator.New()
	if err := valid.Struct(conf); err != nil {
		return errors.Wrap(err, op)
	}
	return nil
}

func SetupLogger() *slog.Logger {
	var envConf EnvConfig
	if err := config.ReadConfig(&envConf); err != nil {
		log.Println(errors.Wrap(err, "setup logger: failed to read config"))
	}

	return SetupLoggerConf(Config{
		Env: envConf.Env,
	})
}

func SetupLoggerConf(conf Config) *slog.Logger {

	if err := validateConfig(conf); err != nil {
		log.Println(err)
	}

	switch conf.Env {

	case EnvProd:
		return slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelInfo},
		))

	case EnvTest:
		return slog.New(slog.NewJSONHandler(
			io.Discard,
			&slog.HandlerOptions{Level: slog.LevelInfo},
		))

	case EnvDev:
		return slog.New(NewHandler(
			WithYamlMarshaller(),
			WithNumbersHighlight(colors.Blue),
			WithKeyWordsHighlight(map[string]colors.ColorCode{
				AttrError:     colors.Red,
				AttrOperation: colors.Green,
			}),
		))
	}

	return slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: slog.LevelDebug},
	))
}
