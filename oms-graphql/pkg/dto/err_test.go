package dto //nolint:testpackage // testing exported API only

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvalidArgument(t *testing.T) {
	t.Parallel()

	err := InvalidArgument("bad input")
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}
