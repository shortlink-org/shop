package dto

import (
	"errors"

	"connectrpc.com/connect"
)

// InvalidArgument returns a Connect error with CodeInvalidArgument.
func InvalidArgument(message string) error {
	return connect.NewError(connect.CodeInvalidArgument, errors.New(message))
}
