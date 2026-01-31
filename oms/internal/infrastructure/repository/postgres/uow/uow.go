package uow

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/tx"
)

// PostgresUoW implements UnitOfWork using PostgreSQL transactions.
type PostgresUoW struct {
	pool *pgxpool.Pool
}

// New creates a new PostgreSQL UnitOfWork.
func New(pool *pgxpool.Pool) *PostgresUoW {
	return &PostgresUoW{pool: pool}
}

// Begin starts a new transaction and returns context with tx.
func (u *PostgresUoW) Begin(ctx context.Context) (context.Context, error) {
	pgxTx, err := u.pool.Begin(ctx)
	if err != nil {
		return ctx, err
	}
	return tx.WithTx(ctx, pgxTx), nil
}

// Commit commits the transaction from context.
func (u *PostgresUoW) Commit(ctx context.Context) error {
	pgxTx := tx.FromContext(ctx)
	if pgxTx == nil {
		return nil // no-op if no transaction
	}
	return pgxTx.Commit(ctx)
}

// Rollback rolls back the transaction from context.
// Safe to call multiple times or after commit (no-op).
func (u *PostgresUoW) Rollback(ctx context.Context) error {
	pgxTx := tx.FromContext(ctx)
	if pgxTx == nil {
		return nil // no-op if no transaction
	}
	return pgxTx.Rollback(ctx)
}
