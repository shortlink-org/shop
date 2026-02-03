package list_by_customer

import (
	"context"
	"fmt"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// Result is the result of the ListOrdersByCustomer query.
type Result = []*orderv1.OrderState

// Handler handles ListOrdersByCustomer queries.
type Handler struct {
	uow       ports.UnitOfWork
	orderRepo ports.OrderRepository
}

// NewHandler creates a new ListOrdersByCustomer handler.
func NewHandler(
	uow ports.UnitOfWork,
	orderRepo ports.OrderRepository,
) (*Handler, error) {
	return &Handler{
		uow:       uow,
		orderRepo: orderRepo,
	}, nil
}

// Handle executes the ListOrdersByCustomer query.
func (h *Handler) Handle(ctx context.Context, q Query) (Result, error) {
	// Begin transaction
	ctx, err := h.uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = h.uow.Rollback(ctx) }()

	orders, err := h.orderRepo.ListByCustomer(ctx, q.CustomerID)
	if err != nil {
		return nil, err
	}

	// Commit transaction (read-only)
	if err := h.uow.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return orders, nil
}

// Ensure Handler implements QueryHandler interface.
var _ ports.QueryHandler[Query, Result] = (*Handler)(nil)
