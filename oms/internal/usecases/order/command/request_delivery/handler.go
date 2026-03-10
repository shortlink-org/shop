package request_delivery

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// Handler handles RequestDelivery commands.
type Handler struct {
	log       logger.Logger
	uow       ports.UnitOfWork
	orderRepo ports.OrderRepository
	publisher ports.EventPublisher
}

// NewHandler creates a new RequestDelivery handler.
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

// Handle executes the RequestDelivery command.
// Pattern: Begin -> Load -> Mutate -> Save -> Publish in tx -> Commit.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	ctx, err := h.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if committed {
			return
		}

		rollbackErr := h.uow.Rollback(ctx)
		if rollbackErr != nil {
			h.log.Warn("transaction rollback failed", slog.Any("error", rollbackErr))
		}
	}()

	order, err := h.orderRepo.Load(ctx, cmd.OrderID)
	if err != nil {
		return fmt.Errorf("failed to load order: %w", err)
	}

	if err := order.RequestDelivery(&cmd.PackageID, cmd.RequestedAt); err != nil {
		return fmt.Errorf("failed to record delivery request: %w", err)
	}

	if err := h.orderRepo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	for _, event := range order.GetDomainEvents() {
		if err := h.publisher.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish domain event to outbox: %w", err)
		}
	}

	if err := h.uow.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true

	order.ClearDomainEvents()

	return nil
}
