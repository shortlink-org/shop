package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
	cart "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/cart/dto"
	"github.com/shortlink-org/shop/oms/pkg/uow"
)

// Load retrieves a cart by customer ID.
// Requires transaction in context (use UnitOfWork.Begin()).
func (s *Store) Load(ctx context.Context, customerID uuid.UUID) (*cart.State, error) {
	pgxTx := uow.FromContext(ctx)
	if pgxTx == nil {
		return nil, ErrTransactionRequired
	}

	qtx := s.query.WithTx(pgxTx)

	// Get cart header
	row, err := qtx.GetCart(ctx, customerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}

	// Get cart items
	items, err := qtx.GetCartItems(ctx, customerID)
	if err != nil {
		return nil, err
	}

	return dto.ToDomain(row, items), nil
}
