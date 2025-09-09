package gocherry

import (
	"bytes"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vishenosik/gocherry/pkg/config"
)

// config.info
// config.gen

type T struct {
	*testing.T
}

func (t *T) Run(name string, f func(t *testing.T)) bool {
	// Cleanup before every test run
	config.Manager().Cleanup()

	return t.T.Run(name, f)
}

func TestFlagsErrors(_t *testing.T) {

	t := &T{_t}

	t.Run("unknown flag", func(t *testing.T) {
		var buf bytes.Buffer
		err := parseFlags(&buf, []string{"-config.unknown"},
			ConfigFlags(&buf, TestConfig{}),
		)
		require.Error(t, err)
		require.NotEmpty(t, buf.String())
	})

	t.Run("help flag", func(t *testing.T) {
		var buf bytes.Buffer
		err := parseFlags(&buf, []string{"-help"},
			ConfigFlags(&buf, TestConfig{}),
		)
		require.ErrorIs(t, err, flag.ErrHelp)
		require.NotEmpty(t, buf.String())
	})
}

func TestConfigFlags(_t *testing.T) {

	t := &T{_t}

	t.Run("config info", func(t *testing.T) {
		var buf bytes.Buffer
		err := parseFlags(&buf, []string{"-config.info"},
			ConfigFlags(&buf, TestConfig{}),
		)
		require.ErrorIs(t, err, ErrSuccessExit)
		require.Equal(t, buf.String(), configInfo)
	})

	t.Run("config gen file", func(t *testing.T) {

		const filename = "config.env"

		file, err := os.CreateTemp("", filename)
		defer os.Remove(file.Name())

		require.NoError(t, err)

		var buf bytes.Buffer
		err = parseFlags(&buf, []string{"-config.gen", file.Name()},
			ConfigFlags(&buf, TestConfig{}),
		)
		require.ErrorIs(t, err, ErrSuccessExit)
		require.Empty(t, buf.String(), configInfo)

		config, err := os.ReadFile(file.Name())
		require.NoError(t, err)
		require.Equal(t, configInfo, string(config))

	})

}

type TestConfig struct {
	Verbose  bool   `env:"VERBOSE" env-default:"true" desc:"Verbose description"`
	Greeting string `env:"GREETING" env-default:"Greeting" desc:"Greeting description"`
	Level    int    `env:"LEVEL" env-default:"123" desc:"Level description"`
}

func (TestConfig) Desc() string {
	return "Test config"
}

var configInfo = `
#=== Test config ===#

# Verbose description (bool)
VERBOSE=true
# Greeting description (string)
GREETING=Greeting
# Level description (int)
LEVEL=123
`
