package ports

import "context"

// CommandHandlerWithResult handles commands that return a result.
// C = Command type, R = Result type.
// Used by calculate_total.Handler and other command handlers.
//
//nolint:iface // interface is implemented and used in other packages
type CommandHandlerWithResult[C any, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}
