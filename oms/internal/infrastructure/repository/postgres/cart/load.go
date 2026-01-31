package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
	cart "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/cart/dto"
)

// Load retrieves a cart by customer ID.
func (s *Store) Load(ctx context.Context, customerID uuid.UUID) (*cart.State, error) {
	// Get cart header
	row, err := s.query.GetCart(ctx, uuidToPgtype(customerID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}

	// Get cart items
	items, err := s.query.GetCartItems(ctx, uuidToPgtype(customerID))
	if err != nil {
		return nil, err
	}

	return dto.ToDomain(row, items), nil
}
