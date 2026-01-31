package tx

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type ctxKey struct{}

// FromContext extracts pgx.Tx from context.
// Returns nil if no transaction in context.
func FromContext(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(ctxKey{}).(pgx.Tx)
	return tx
}

// WithTx returns context with pgx.Tx embedded.
func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, ctxKey{}, tx)
}

// HasTx checks if context has a transaction.
func HasTx(ctx context.Context) bool {
	return FromContext(ctx) != nil
}
