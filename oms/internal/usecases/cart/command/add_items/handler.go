package add_items

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/oms/internal/domain"
	v1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// Handler handles AddItems commands.
type Handler struct {
	log       logger.Logger
	uow       ports.UnitOfWork
	cartRepo  ports.CartRepository
	publisher ports.EventPublisher
}

// NewHandler creates a new AddItems handler.
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

// Handle executes the AddItems command.
// Pattern: Load -> Domain method -> Save
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	// Begin transaction
	ctx, err := h.uow.Begin(ctx)
	if err != nil {
		return domain.MapInfraErr("uow.Begin", err)
	}

	defer func() {
		rollbackErr := h.uow.Rollback(ctx)
		if rollbackErr != nil {
			h.log.Warn("transaction rollback failed", slog.Any("error", rollbackErr))
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

	// 2. Call domain method for each item
	for _, item := range cmd.Items {
		addErr := cart.AddItem(item)
		if addErr != nil {
			return domain.WrapValidation("cart.AddItem", addErr)
		}
	}

	// 3. Save aggregate
	if err := h.cartRepo.Save(ctx, cart); err != nil {
		return domain.MapInfraErr("cartRepo.Save", err)
	}

	// 4. Publish domain events to outbox (same transaction)
	for _, event := range cart.GetDomainEvents() {
		pubErr := h.publisher.Publish(ctx, event)
		if pubErr != nil {
			return domain.MapInfraErr("eventBus.Publish", pubErr)
		}
	}

	cart.ClearDomainEvents()

	// 5. Commit transaction
	if err := h.uow.Commit(ctx); err != nil {
		return domain.MapInfraErr("uow.Commit", err)
	}

	return nil
}
