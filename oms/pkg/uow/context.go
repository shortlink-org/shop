package uow

import (
	"context"

	"github.com/jackc/pgx/v5"
	sdkuow "github.com/shortlink-org/go-sdk/uow"
)

// Re-export go-sdk/uow so OMS and cqrs share the same context key for pgx.Tx.

// FromContext returns the pgx.Tx from ctx, or nil if not set.
func FromContext(ctx context.Context) pgx.Tx {
	return sdkuow.FromContext(ctx)
}

// WithTx returns a context that carries the given transaction.
func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return sdkuow.WithTx(ctx, tx)
}

// HasTx reports whether ctx contains a transaction.
func HasTx(ctx context.Context) bool {
	return sdkuow.HasTx(ctx)
}

// ContextWithoutTx returns a context that has the same deadline as ctx (if any)
// but does not carry the transaction. Use it for operations that must run after
// Commit (e.g. updating an external index) to avoid using a closed tx.
// The caller should call the returned CancelFunc when done (e.g. defer cancel()).
func ContextWithoutTx(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx2 := context.Background()
	if deadline, ok := ctx.Deadline(); ok {
		return context.WithDeadline(ctx2, deadline)
	}

	return context.WithCancel(ctx2)
}
