package list

import (
	"context"
	"fmt"
	"log/slog"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// Result is the result of the ListOrders query (slice of aggregates).
type Result = []*order.OrderState

// Handler handles ListOrders queries.
type Handler struct {
	uow       ports.UnitOfWork
	orderRepo ports.OrderRepository
}

// NewHandler creates a new ListOrders handler.
func NewHandler(
	uow ports.UnitOfWork,
	orderRepo ports.OrderRepository,
) (*Handler, error) {
	return &Handler{
		uow:       uow,
		orderRepo: orderRepo,
	}, nil
}

// Handle executes the ListOrders query.
func (h *Handler) Handle(ctx context.Context, q Query) (Result, error) {
	ctx, err := h.uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		err := h.uow.Rollback(ctx)
		if err != nil {
			slog.Default().WarnContext(ctx, "transaction rollback failed", "error", err)
		}
	}()

	filter := ports.ListFilter{
		CustomerID:   q.CustomerID,
		StatusFilter: q.StatusFilter,
	}

	orders, err := h.orderRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err := h.uow.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return orders, nil
}
