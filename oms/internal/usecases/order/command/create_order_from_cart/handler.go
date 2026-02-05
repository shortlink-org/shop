package create_order_from_cart

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/shopspring/decimal"
	"github.com/shortlink-org/go-sdk/logger"

	orderDomain "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

var (
	errEmptyCart            = errors.New("cannot create order from empty cart")
	errInvalidDeliveryInfo  = errors.New("invalid delivery info")
	errPricerClientRequired = errors.New("pricer client required for checkout")
)

// Result represents the result of creating an order from a cart.
type Result struct {
	Order         *orderDomain.OrderState
	Subtotal      decimal.Decimal
	TotalDiscount decimal.Decimal
	TotalTax      decimal.Decimal
	FinalPrice    decimal.Decimal
}

// Handler handles CreateOrderFromCart commands.
type Handler struct {
	log          logger.Logger
	uow          ports.UnitOfWork
	cartRepo     ports.CartRepository
	orderRepo    ports.OrderRepository
	publisher    ports.EventPublisher
	pricerClient ports.PricerClient
}

// NewHandler creates a new CreateOrderFromCart handler.
func NewHandler(
	log logger.Logger,
	uow ports.UnitOfWork,
	cartRepo ports.CartRepository,
	orderRepo ports.OrderRepository,
	publisher ports.EventPublisher,
	pricerClient ports.PricerClient,
) (*Handler, error) {
	return &Handler{
		log:          log,
		uow:          uow,
		cartRepo:     cartRepo,
		orderRepo:    orderRepo,
		publisher:    publisher,
		pricerClient: pricerClient,
	}, nil
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
		err := h.uow.Rollback(ctx)
		if err != nil {
			h.log.Warn("rollback failed", slog.Any("error", err))
		}
	}()

	// 2. Load cart (uses tx from ctx)
	cart, err := h.cartRepo.Load(ctx, cmd.CustomerID)
	if err != nil {
		return Result{}, fmt.Errorf("failed to load cart: %w", err)
	}

	// 3. Validate cart is not empty
	cartItems := cart.GetItems()
	if len(cartItems) == 0 {
		return Result{}, errEmptyCart
	}

	// 4. Validate delivery info if provided
	if cmd.DeliveryInfo != nil && !cmd.DeliveryInfo.IsValid() {
		return Result{}, errInvalidDeliveryInfo
	}

	// 5. Calculate pricing using Pricer service (required for taxes and correct charge)
	if h.pricerClient == nil {
		return Result{}, errPricerClientRequired
	}

	builder := NewPricerRequestBuilder(cart.GetCustomerId(), cartItems)

	if cmd.DeliveryInfo != nil {
		addr := cmd.DeliveryInfo.GetDeliveryAddress()
		builder = builder.
			WithTaxParam("country", addr.Country()).
			WithTaxParam("city", addr.City()).
			WithTaxParam("postalCode", addr.PostalCode())
	}

	pricingResp, err := h.pricerClient.CalculateTotal(ctx, builder.Build())
	if err != nil {
		return Result{}, fmt.Errorf("failed to calculate pricing: %w", err)
	}

	// 6. Prepare neutral lines from cart (application-layer mapping)
	lines := cartItemsToLines(cartItems)

	// 7. Create order from lines (domain keeps invariants)
	order := orderDomain.NewOrderState(cmd.CustomerID)
	if err := order.CreateFromLines(ctx, lines); err != nil {
		return Result{}, fmt.Errorf("failed to create order: %w", err)
	}

	// 8. Set delivery info if provided
	if cmd.DeliveryInfo != nil {
		err := order.SetDeliveryInfo(*cmd.DeliveryInfo)
		if err != nil {
			return Result{}, fmt.Errorf("failed to set delivery info: %w", err)
		}
	}

	// 9. Clear cart
	cart.Reset()

	// 10. Save order (uses tx from ctx)
	if err := h.orderRepo.Save(ctx, order); err != nil {
		return Result{}, fmt.Errorf("failed to save order: %w", err)
	}

	// 11. Save cart (uses tx from ctx)
	if err := h.cartRepo.Save(ctx, cart); err != nil {
		return Result{}, fmt.Errorf("failed to save cart: %w", err)
	}

	// 12. Publish domain events to outbox (same transaction).
	// If outbox write fails, we must not commit — same as failing to save order/cart.
	for _, event := range order.GetDomainEvents() {
		err := h.publisher.Publish(ctx, event)
		if err != nil {
			return Result{}, fmt.Errorf("failed to publish domain event to outbox: %w", err)
		}
	}

	order.ClearDomainEvents()

	// 13. Commit transaction
	if err := h.uow.Commit(ctx); err != nil {
		return Result{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 14. Build result with pricing info
	return Result{
		Order:         order,
		Subtotal:      pricingResp.Subtotal,
		TotalDiscount: pricingResp.TotalDiscount,
		TotalTax:      pricingResp.TotalTax,
		FinalPrice:    pricingResp.FinalPrice,
	}, nil
}
