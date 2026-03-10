package dto

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
)

func TestInvalidArgument(t *testing.T) {
	t.Parallel()
	err := InvalidArgument("bad input")
	assert.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}
