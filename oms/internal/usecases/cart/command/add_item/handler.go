package add_item

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/oms/internal/domain"
	v1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// Handler handles AddItem commands.
type Handler struct {
	log       logger.Logger
	uow       ports.UnitOfWork
	cartRepo  ports.CartRepository
	publisher ports.EventPublisher
}

// NewHandler creates a new AddItem handler.
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

// Handle executes the AddItem command.
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

	// 1. Load aggregate (or create new if not found)
	cart, err := h.cartRepo.Load(ctx, cmd.CustomerID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			cart = v1.New(cmd.CustomerID)
		} else {
			return domain.MapInfraErr("cartRepo.Load", err)
		}
	}

	// 2. Call domain method (business logic)
	err = cart.AddItem(cmd.Item)
	if err != nil {
		return domain.WrapValidation("cart.AddItem", err)
	}

	// 3. Save aggregate
	err = h.cartRepo.Save(ctx, cart)
	if err != nil {
		return domain.MapInfraErr("cartRepo.Save", err)
	}

	// 4. Publish domain events to outbox (same transaction)
	for _, event := range cart.GetDomainEvents() {
		err = h.publisher.Publish(ctx, event)
		if err != nil {
			return domain.MapInfraErr("eventBus.Publish", err)
		}
	}

	cart.ClearDomainEvents()

	// 5. Commit transaction
	err = h.uow.Commit(ctx)
	if err != nil {
		return domain.MapInfraErr("uow.Commit", err)
	}

	return nil
}
