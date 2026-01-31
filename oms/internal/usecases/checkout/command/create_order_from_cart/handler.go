package create_order_from_cart

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	cartItemsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/items/v1"
	orderDomain "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// Result represents the result of creating an order from a cart.
type Result struct {
	Order *orderDomain.OrderState
}

// Handler handles CreateOrderFromCart commands.
type Handler struct {
	log       logger.Logger
	uow       ports.UnitOfWork
	cartRepo  ports.CartRepository
	orderRepo ports.OrderRepository
	publisher ports.EventPublisher
}

// NewHandler creates a new CreateOrderFromCart handler.
func NewHandler(
	log logger.Logger,
	uow ports.UnitOfWork,
	cartRepo ports.CartRepository,
	orderRepo ports.OrderRepository,
	publisher ports.EventPublisher,
) *Handler {
	return &Handler{
		log:       log,
		uow:       uow,
		cartRepo:  cartRepo,
		orderRepo: orderRepo,
		publisher: publisher,
	}
}

// Handle executes the CreateOrderFromCart command.
// Atomically creates an order from cart and clears cart.
func (h *Handler) Handle(ctx context.Context, cmd Command) (Result, error) {
	// 1. Begin transaction
	ctx, err := h.uow.Begin(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		// Rollback is no-op if already committed
		_ = h.uow.Rollback(ctx)
	}()

	// 2. Load cart (uses tx from ctx)
	cart, err := h.cartRepo.Load(ctx, cmd.CustomerID)
	if err != nil {
		return Result{}, fmt.Errorf("failed to load cart: %w", err)
	}

	// 3. Validate cart is not empty
	cartItems := cart.GetItems()
	if len(cartItems) == 0 {
		return Result{}, fmt.Errorf("cannot create order from empty cart")
	}

	// 4. Validate delivery info if provided
	if cmd.DeliveryInfo != nil && !cmd.DeliveryInfo.IsValid() {
		return Result{}, fmt.Errorf("invalid delivery info")
	}

	// 5. Convert cart items to order items
	orderItems := convertCartToOrderItems(cartItems)

	// 6. Create order from cart items
	order := orderDomain.NewOrderState(cmd.CustomerID)
	if err := order.CreateOrder(orderItems); err != nil {
		return Result{}, fmt.Errorf("failed to create order: %w", err)
	}

	// 7. Set delivery info if provided
	if cmd.DeliveryInfo != nil {
		if err := order.SetDeliveryInfo(*cmd.DeliveryInfo); err != nil {
			return Result{}, fmt.Errorf("failed to set delivery info: %w", err)
		}
	}

	// 8. Clear cart
	cart.Reset()

	// 9. Save order (uses tx from ctx)
	if err := h.orderRepo.Save(ctx, order); err != nil {
		return Result{}, fmt.Errorf("failed to save order: %w", err)
	}

	// 10. Save cart (uses tx from ctx)
	if err := h.cartRepo.Save(ctx, cart); err != nil {
		return Result{}, fmt.Errorf("failed to save cart: %w", err)
	}

	// 11. Commit transaction
	if err := h.uow.Commit(ctx); err != nil {
		return Result{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 12. Publish domain events
	for _, event := range order.GetDomainEvents() {
		if publishableEvent, ok := event.(ports.Event); ok {
			if err := h.publisher.Publish(ctx, publishableEvent); err != nil {
				h.log.Error("failed to publish domain event",
					slog.String("event_type", event.EventType()),
					slog.String("order_id", order.GetOrderID().String()),
					slog.Any("error", err))
			}
		}
	}
	order.ClearDomainEvents()

	return Result{Order: order}, nil
}

// convertCartToOrderItems converts cart items to order items.
func convertCartToOrderItems(cartItems cartItemsv1.Items) orderDomain.Items {
	orderItems := make(orderDomain.Items, 0, len(cartItems))
	for _, cartItem := range cartItems {
		// Use price from cart item (already includes discount calculation)
		orderItem := orderDomain.NewItem(
			cartItem.GetGoodId(),
			cartItem.GetQuantity(),
			cartItem.GetPrice(),
		)
		orderItems = append(orderItems, orderItem)
	}
	return orderItems
}

// Ensure Handler implements CommandHandlerWithResult interface.
var _ ports.CommandHandlerWithResult[Command, Result] = (*Handler)(nil)
