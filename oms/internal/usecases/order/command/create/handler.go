package create

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shortlink-org/go-sdk/logger"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// Handler handles CreateOrder commands.
type Handler struct {
	log       logger.Logger
	uow       ports.UnitOfWork
	orderRepo ports.OrderRepository
	publisher ports.EventPublisher
}

// NewHandler creates a new CreateOrder handler.
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

// Handle executes the CreateOrder command.
// Pattern: Create aggregate -> Persist -> Publish event
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	// Begin transaction
	ctx, err := h.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = h.uow.Rollback(ctx) }()

	// 1. Create domain aggregate
	order := orderv1.NewOrderState(cmd.CustomerID)
	order.SetID(cmd.OrderID)

	// 2. Apply business logic (create order with items)
	if err := order.CreateOrder(cmd.Items); err != nil {
		return err
	}

	// 3. Set delivery info if provided
	if cmd.DeliveryInfo != nil {
		if err := order.SetDeliveryInfo(*cmd.DeliveryInfo); err != nil {
			return fmt.Errorf("failed to set delivery info: %w", err)
		}
	}

	// 4. Persist to database
	if err := h.orderRepo.Save(ctx, order); err != nil {
		return err
	}

	// 5. Publish domain events to outbox (same transaction)
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

// Ensure Handler implements CommandHandler interface.
var _ ports.CommandHandler[Command] = (*Handler)(nil)
