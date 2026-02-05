package remove_items

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/oms/internal/domain"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// Handler handles RemoveItems commands.
type Handler struct {
	log       logger.Logger
	uow       ports.UnitOfWork
	cartRepo  ports.CartRepository
	publisher ports.EventPublisher
}

// NewHandler creates a new RemoveItems handler.
func NewHandler(
	log logger.Logger,
	uow ports.UnitOfWork,
	cartRepo ports.CartRepository,
	publisher ports.EventPublisher,
) (*Handler, error) {
	return &Handler{
		log:       log,
		uow:       uow,
		cartRepo:  cartRepo,
		publisher: publisher,
	}, nil
}

// Handle executes the RemoveItems command.
// Pattern: Load -> Domain method -> Save
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	// Begin transaction
	ctx, err := h.uow.Begin(ctx)
	if err != nil {
		return domain.MapInfraErr("uow.Begin", err)
	}

	defer func() {
		err := h.uow.Rollback(ctx)
		if err != nil {
			h.log.Warn("transaction rollback failed", slog.Any("error", err))
		}
	}()

	// 1. Load aggregate
	cart, err := h.cartRepo.Load(ctx, cmd.CustomerID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}

		return domain.MapInfraErr("cartRepo.Load", err)
	}

	// 2. Call domain method for each item
	for _, item := range cmd.Items {
		err := cart.RemoveItem(item)
		if err != nil {
			return domain.WrapValidation("cart.RemoveItem", err)
		}
	}

	// 3. Save aggregate
	if err := h.cartRepo.Save(ctx, cart); err != nil {
		return domain.MapInfraErr("cartRepo.Save", err)
	}

	// 4. Publish domain events to outbox (same transaction)
	for _, event := range cart.GetDomainEvents() {
		err := h.publisher.Publish(ctx, event)
		if err != nil {
			return domain.MapInfraErr("eventBus.Publish", err)
		}
	}

	cart.ClearDomainEvents()

	// 5. Commit transaction
	if err := h.uow.Commit(ctx); err != nil {
		return domain.MapInfraErr("uow.Commit", err)
	}

	return nil
}
