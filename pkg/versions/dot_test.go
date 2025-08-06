package versions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {

	v1 := NewDotVersion("1.0")
	v2 := NewDotVersion("1.0")
	v3 := NewDotVersion("1.0")
	v4 := NewDotVersion("1.1")

	require.True(t, v3.In(v1, v2))

	require.False(t, v4.In(v1, v2))
}
