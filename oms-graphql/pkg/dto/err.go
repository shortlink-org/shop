package dto

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"
)

// ErrInvalidArgument is the sentinel for invalid argument errors.
var ErrInvalidArgument = errors.New("invalid argument")

// InvalidArgument returns a Connect error with CodeInvalidArgument.
func InvalidArgument(message string) error {
	return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%w: %s", ErrInvalidArgument, message))
}
