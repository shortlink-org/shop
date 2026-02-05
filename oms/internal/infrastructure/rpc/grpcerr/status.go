package grpcerr

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/shortlink-org/shop/oms/internal/domain"
)

// Logger is the minimal interface for logging unexpected errors (Op + Unwrap).
type Logger interface {
	Warn(string, ...slog.Attr)
}

// ToStatus maps a usecase/domain error to a gRPC status and logs Op + Unwrap for diagnostics.
// Returns the gRPC error to return from the handler.
func ToStatus(ctx context.Context, log Logger, op string, err error) error {
	if err == nil {
		return nil
	}

	var (
		code codes.Code
		msg  string
	)

	switch {
	case errors.Is(err, domain.ErrValidation):
		code = codes.InvalidArgument
		msg = err.Error()
	case errors.Is(err, domain.ErrNotFound):
		code = codes.NotFound
		msg = err.Error()
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrVersionConflict):
		code = codes.Aborted
		msg = err.Error()
	case errors.Is(err, domain.ErrUnavailable):
		code = codes.Unavailable
		msg = err.Error()
	default:
		code = codes.Internal
		msg = "internal error"

		if log != nil {
			log.Warn("rpc error", slog.String("op", op), slog.Any("error", err), slog.Any("unwrap", errors.Unwrap(err)))
		}
	}

	return status.Error(code, msg)
}
