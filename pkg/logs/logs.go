package logs

import (
	"io"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/vishenosik/gocherry/pkg/colors"
	"github.com/vishenosik/gocherry/pkg/config"
)

const (
	EnvDev  = "dev"
	EnvProd = "prod"
	EnvTest = "test"
)

func init() {
	config.AddStructs(EnvConfig{})
}

type EnvConfig struct {
	Env string `env:"ENV" env-default:"dev" desc:"The environment in which the application is running"`
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
	if err := config.ReadConfigEnv(&envConf); err != nil {
		// log.Println(errors.Wrap(err, "setup logger: failed to read config"))
	}

	return SetupLoggerConf(Config{
		Env: envConf.Env,
	})
}

func SetupLoggerConf(conf Config) *slog.Logger {

	if err := validateConfig(conf); err != nil {
		// log.Println(err)
	}

	var handler slog.Handler

	switch conf.Env {

	case EnvProd:
		handler = slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelInfo},
		)

	case EnvTest:
		handler = slog.NewJSONHandler(
			io.Discard,
			&slog.HandlerOptions{Level: slog.LevelInfo},
		)

	case EnvDev:
		handler = NewHandler(
			WithYamlMarshaller(),
			WithNumbersHighlight(colors.Blue),
			WithKeyWordsHighlight(map[string]colors.ColorCode{
				AttrError:     colors.Red,
				AttrOperation: colors.Green,
			}),
		)

	default:
		handler = slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug},
		)
	}

	logger := slog.New(handler)
	redirectStdLogger(logger)

	return logger
}
