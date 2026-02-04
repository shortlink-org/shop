package get

import (
	"context"
	"fmt"
	"log/slog"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// Result is the result of the GetOrder query.
type Result = *orderv1.OrderState

// Handler handles GetOrder queries.
type Handler struct {
	uow       ports.UnitOfWork
	orderRepo ports.OrderRepository
}

// NewHandler creates a new GetOrder handler.
func NewHandler(
	uow ports.UnitOfWork,
	orderRepo ports.OrderRepository,
) (*Handler, error) {
	return &Handler{
		uow:       uow,
		orderRepo: orderRepo,
	}, nil
}

// Handle executes the GetOrder query.
func (h *Handler) Handle(ctx context.Context, q Query) (Result, error) {
	// Begin transaction
	ctx, err := h.uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := h.uow.Rollback(ctx); err != nil {
			slog.Default().WarnContext(ctx, "transaction rollback failed", "error", err)
		}
	}()

	order, err := h.orderRepo.Load(ctx, q.OrderID)
	if err != nil {
		return nil, err
	}

	// Commit transaction (read-only)
	if err := h.uow.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return order, nil
}
