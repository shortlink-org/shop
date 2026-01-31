package ports

import "context"

// UnitOfWork manages transaction lifecycle.
// It does NOT know about repositories — only about transactions.
// Repositories detect transaction in context and participate automatically.
type UnitOfWork interface {
	// Begin starts a new transaction and returns context with tx.
	Begin(ctx context.Context) (context.Context, error)

	// Commit commits the transaction from context.
	Commit(ctx context.Context) error

	// Rollback rolls back the transaction from context.
	// Safe to call multiple times or after commit (no-op).
	Rollback(ctx context.Context) error
}
