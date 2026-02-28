package ports

import "context"

// UnitOfWork manages transaction lifecycle.
// It does NOT know about repositories — only about transactions.
// Repositories detect transaction in context and participate automatically.
type UnitOfWork interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
