package list

import (
	"context"
	"fmt"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// Result is the result of the ListOrders query.
type Result = *ports.ListResult

// Handler handles ListOrders queries.
type Handler struct {
	uow       ports.UnitOfWork
	orderRepo ports.OrderRepository
}

// NewHandler creates a new ListOrders handler.
func NewHandler(
	uow ports.UnitOfWork,
	orderRepo ports.OrderRepository,
) *Handler {
	return &Handler{
		uow:       uow,
		orderRepo: orderRepo,
	}
}

// Handle executes the ListOrders query.
func (h *Handler) Handle(ctx context.Context, q Query) (Result, error) {
	// Begin transaction
	ctx, err := h.uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = h.uow.Rollback(ctx) }()

	filter := ports.ListFilter{
		CustomerID:   q.CustomerID,
		StatusFilter: q.StatusFilter,
	}

	result, err := h.orderRepo.List(ctx, filter, q.Page, q.PageSize)
	if err != nil {
		return nil, err
	}

	// Commit transaction (read-only)
	if err := h.uow.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// Ensure Handler implements QueryHandler interface.
var _ ports.QueryHandler[Query, Result] = (*Handler)(nil)
