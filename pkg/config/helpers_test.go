package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

type T struct {
	*testing.T
}

func (t *T) Run(name string, f func(t *testing.T)) bool {
	// Cleanup before every test run
	Manager().Cleanup()

	return t.T.Run(name, f)
}

func TestConfigInfoEnv(_t *testing.T) {

	t := &T{_t}

	t.Run("Straight register", func(t *testing.T) {
		var buf bytes.Buffer
		ConfigInfoEnv(&buf, TestConfig{})
		require.Equal(t, configInfoTestStraight, buf.String())
	})

	t.Run("Buffer register", func(t *testing.T) {
		var buf bytes.Buffer
		AddStructs(TestConfig{})
		ConfigInfoEnv(&buf)
		require.Equal(t, configInfoTestStraight, buf.String())
	})

	t.Run("Combined register", func(t *testing.T) {
		var buf bytes.Buffer
		AddStructs(TestConfig{})
		ConfigInfoEnv(&buf, TestConfig{})
		require.Equal(t, configInfoTestCombined, buf.String())
	})

	t.Run("Empty register", func(t *testing.T) {
		var buf bytes.Buffer
		ConfigInfoEnv(&buf)
		require.Equal(t, "", buf.String())
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

var configInfoTestStraight = `
#=== Test config ===#

# Verbose description (bool)
VERBOSE=true
# Greeting description (string)
GREETING=Greeting
# Level description (int)
LEVEL=123
`

var configInfoTestCombined = `
#=== Test config ===#

# Verbose description (bool)
VERBOSE=true
# Greeting description (string)
GREETING=Greeting
# Level description (int)
LEVEL=123

#=== Test config ===#

# Verbose description (bool)
VERBOSE=true
# Greeting description (string)
GREETING=Greeting
# Level description (int)
LEVEL=123
`
