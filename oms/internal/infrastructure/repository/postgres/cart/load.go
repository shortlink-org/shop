package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/shortlink-org/shop/oms/internal/domain"
	cart "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/cart/dto"
	"github.com/shortlink-org/shop/oms/pkg/uow"
)

func cloneCartState(state *cart.State) *cart.State {
	if state == nil {
		return nil
	}

	return cart.Reconstitute(state.GetCustomerId(), state.GetItems(), state.GetVersion())
}

// Load retrieves a cart by customer ID.
// Uses L1 cache for frequently accessed carts.
// Requires transaction in context (use UnitOfWork.Begin()).
func (s *Store) Load(ctx context.Context, customerID uuid.UUID) (*cart.State, error) {
	// Check L1 cache first
	cacheKey := customerID.String()
	if cachedCart, found := s.cache.Get(cacheKey); found {
		return cloneCartState(cachedCart), nil
	}

	// Cache miss - fetch from database
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

		return nil, domain.WrapUnavailable("GetCart", err)
	}

	// Get cart items
	items, err := qtx.GetCartItems(ctx, customerID)
	if err != nil {
		return nil, domain.WrapUnavailable("GetCartItems", err)
	}

	result := dto.ToDomain(row, items)

	// Store in L1 cache: cost = base + items * per-item cost
	cost := int64(100 + len(items)*50)
	s.cache.SetWithTTL(cacheKey, cloneCartState(result), cost, cacheTTL)

	return result, nil
}
