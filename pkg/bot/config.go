package bot

import "github.com/vishenosik/gocherry/pkg/errors"

type Config struct {
	Token   string
	Timeout int
}

func NewConfig(token string) Config {
	return Config{
		Token:   token,
		Timeout: 10,
	}
}

func (c Config) validate() error {
	errs := &errors.MultiError{}

	if c.Token == "" {
		errs.Append(errors.New("token is empty"))
	}

	if c.Timeout <= 0 {
		errs.Append(errors.New("timeout must be greater than 0"))
	}

	return errs.ErrorOrNil()
}

type ConfigEnv struct {
	Token   string `env:"TG_BOT_TOKEN"`
	Timeout int    `env:"TG_BOT_TIMEOUT" env-default:"10"`
}

func (c ConfigEnv) ToConfig() Config {
	return Config{
		Token:   c.Token,
		Timeout: c.Timeout,
	}
}
