package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	v1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// Get retrieves an order by ID from the database.
func (uc *UC) Get(ctx context.Context, orderID uuid.UUID) (*v1.OrderState, error) {
	// Begin transaction
	ctx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = uc.uow.Rollback(ctx) }()

	order, err := uc.orderRepo.Load(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Commit transaction (read-only)
	if err := uc.uow.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return order, nil
}

// ListByCustomer retrieves all orders for a customer.
func (uc *UC) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*v1.OrderState, error) {
	// Begin transaction
	ctx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = uc.uow.Rollback(ctx) }()

	orders, err := uc.orderRepo.ListByCustomer(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// Commit transaction (read-only)
	if err := uc.uow.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return orders, nil
}
