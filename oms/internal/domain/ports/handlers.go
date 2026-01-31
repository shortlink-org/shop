package ports

import "context"

// CommandHandler handles commands that modify state.
// C = Command type
type CommandHandler[C any] interface {
	Handle(ctx context.Context, cmd C) error
}

// CommandHandlerWithResult handles commands that return a result.
// C = Command type, R = Result type
type CommandHandlerWithResult[C any, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

// QueryHandler handles read-only queries.
// Q = Query type, R = Result type
type QueryHandler[Q any, R any] interface {
	Handle(ctx context.Context, q Q) (R, error)
}

// EventHandler handles domain events (reactions to facts).
// E = Event type
type EventHandler[E any] interface {
	Handle(ctx context.Context, event E) error
}
