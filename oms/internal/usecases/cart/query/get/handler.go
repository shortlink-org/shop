package get

import (
	"context"
	"errors"
	"fmt"

	v1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// Result is the result of the GetCart query.
type Result = *v1.State

// Handler handles GetCart queries.
type Handler struct {
	uow      ports.UnitOfWork
	cartRepo ports.CartRepository
}

// NewHandler creates a new GetCart handler.
func NewHandler(
	uow ports.UnitOfWork,
	cartRepo ports.CartRepository,
) (*Handler, error) {
	return &Handler{
		uow:      uow,
		cartRepo: cartRepo,
	}, nil
}

// Handle executes the GetCart query.
func (h *Handler) Handle(ctx context.Context, q Query) (Result, error) {
	// Begin transaction
	ctx, err := h.uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = h.uow.Rollback(ctx) }()

	cart, err := h.cartRepo.Load(ctx, q.CustomerID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			// Return empty cart if not found
			return v1.New(q.CustomerID), nil
		}
		return nil, err
	}

	// Commit transaction (read-only, but still needs to close tx)
	if err := h.uow.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return cart, nil
}
