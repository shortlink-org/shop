package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shortlink-org/shop/oms/pkg/uow"
)

// UoW implements UnitOfWork using PostgreSQL transactions.
type UoW struct {
	pool *pgxpool.Pool
}

// New creates a new PostgreSQL UnitOfWork.
func New(pool *pgxpool.Pool) *UoW {
	return &UoW{pool: pool}
}

// Begin starts a new transaction and returns context with tx.
func (u *UoW) Begin(ctx context.Context) (context.Context, error) {
	pgxTx, err := u.pool.Begin(ctx)
	if err != nil {
		return ctx, err
	}

	return uow.WithTx(ctx, pgxTx), nil
}

// Commit commits the transaction from context.
func (u *UoW) Commit(ctx context.Context) error {
	pgxTx := uow.FromContext(ctx)
	if pgxTx == nil {
		return nil // no-op if no transaction
	}

	return pgxTx.Commit(ctx)
}

// Rollback rolls back the transaction from context.
// Safe to call multiple times or after commit (no-op).
func (u *UoW) Rollback(ctx context.Context) error {
	pgxTx := uow.FromContext(ctx)
	if pgxTx == nil {
		return nil // no-op if no transaction
	}

	return pgxTx.Rollback(ctx)
}
