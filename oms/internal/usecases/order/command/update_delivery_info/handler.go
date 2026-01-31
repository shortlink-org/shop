package update_delivery_info

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// Handler handles UpdateDeliveryInfo commands.
type Handler struct {
	log       logger.Logger
	uow       ports.UnitOfWork
	orderRepo ports.OrderRepository
	publisher ports.EventPublisher
}

// NewHandler creates a new UpdateDeliveryInfo handler.
func NewHandler(
	log logger.Logger,
	uow ports.UnitOfWork,
	orderRepo ports.OrderRepository,
	publisher ports.EventPublisher,
) *Handler {
	return &Handler{
		log:       log,
		uow:       uow,
		orderRepo: orderRepo,
		publisher: publisher,
	}
}

// Handle executes the UpdateDeliveryInfo command.
// Pattern: Load -> Domain method -> Save -> Publish event
// Returns error if order is in terminal state or delivery is already in progress.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	// Begin transaction
	ctx, err := h.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = h.uow.Rollback(ctx) }()

	// 1. Load order aggregate
	order, err := h.orderRepo.Load(ctx, cmd.OrderID)
	if err != nil {
		return fmt.Errorf("failed to load order: %w", err)
	}

	// 2. Apply business logic (update delivery info)
	// This will fail if order is COMPLETED/CANCELLED or if delivery is ASSIGNED/IN_TRANSIT/DELIVERED
	if err := order.SetDeliveryInfo(cmd.DeliveryInfo); err != nil {
		return fmt.Errorf("cannot update delivery info: %w", err)
	}

	// 3. Persist to database
	if err := h.orderRepo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Commit transaction before publishing events
	if err := h.uow.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 4. Publish domain events (if any)
	for _, event := range order.GetDomainEvents() {
		if publishableEvent, ok := event.(ports.Event); ok {
			if err := h.publisher.Publish(ctx, publishableEvent); err != nil {
				// Log error but don't fail - order is already persisted
				h.log.Error("failed to publish domain event",
					slog.String("event_type", event.EventType()),
					slog.String("order_id", cmd.OrderID.String()),
					slog.Any("error", err))
			}
		}
	}
	order.ClearDomainEvents()

	return nil
}

// Ensure Handler implements CommandHandler interface.
var _ ports.CommandHandler[Command] = (*Handler)(nil)
