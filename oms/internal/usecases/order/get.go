package order

import (
	"context"

	"github.com/google/uuid"

	v1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// Get retrieves an order by ID from the database.
func (uc *UC) Get(ctx context.Context, orderID uuid.UUID) (*v1.OrderState, error) {
	return uc.orderRepo.Load(ctx, orderID)
}

// ListByCustomer retrieves all orders for a customer.
func (uc *UC) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*v1.OrderState, error) {
	return uc.orderRepo.ListByCustomer(ctx, customerID)
}
