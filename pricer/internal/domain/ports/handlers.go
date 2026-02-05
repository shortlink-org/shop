package ports

import "context"

// CommandHandlerWithResult handles commands that return a result.
// C = Command type, R = Result type.
type CommandHandlerWithResult[C any, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}
