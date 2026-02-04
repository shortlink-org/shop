package cancel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// Handler handles CancelOrder commands.
type Handler struct {
	log       logger.Logger
	uow       ports.UnitOfWork
	orderRepo ports.OrderRepository
	publisher ports.EventPublisher
}

// NewHandler creates a new CancelOrder handler.
func NewHandler(
	log logger.Logger,
	uow ports.UnitOfWork,
	orderRepo ports.OrderRepository,
	publisher ports.EventPublisher,
) (*Handler, error) {
	return &Handler{
		log:       log,
		uow:       uow,
		orderRepo: orderRepo,
		publisher: publisher,
	}, nil
}

// Handle executes the CancelOrder command.
// Pattern: Load -> Domain method -> Save -> Publish event
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	// Begin transaction
	ctx, err := h.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := h.uow.Rollback(ctx); err != nil {
			h.log.Warn("transaction rollback failed", slog.Any("error", err))
		}
	}()

	// 1. Load order aggregate
	order, err := h.orderRepo.Load(ctx, cmd.OrderID)
	if err != nil {
		return err
	}

	// 2. Apply business logic (cancel order)
	if err := order.CancelOrder(); err != nil {
		return err
	}

	// 3. Persist to database
	if err := h.orderRepo.Save(ctx, order); err != nil {
		return err
	}

	// 4. Publish domain events to outbox (same transaction)
	for _, event := range order.GetDomainEvents() {
		if err := h.publisher.Publish(ctx, event); err != nil {
			h.log.Error("failed to publish domain event",
				slog.String("order_id", cmd.OrderID.String()),
				slog.Any("error", err))
		}
	}

	if err := h.uow.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	order.ClearDomainEvents()

	return nil
}
