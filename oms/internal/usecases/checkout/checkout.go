package checkout

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	cartItemsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/items/v1"
	orderDomain "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// CreateOrderFromCartResult represents the result of creating an order from a cart
type CreateOrderFromCartResult struct {
	Order *orderDomain.OrderState
}

// CreateOrderFromCart atomically creates an order from cart and clears cart.
// Uses UnitOfWork for transactional consistency across both aggregates.
// deliveryInfo is optional (nil = self-pickup).
func (uc *UC) CreateOrderFromCart(ctx context.Context, customerID uuid.UUID, deliveryInfo *orderDomain.DeliveryInfo) (*CreateOrderFromCartResult, error) {
	// 1. Begin transaction
	ctx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		// Rollback is no-op if already committed
		_ = uc.uow.Rollback(ctx)
	}()

	// 2. Load cart (uses tx from ctx)
	cart, err := uc.cartRepo.Load(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to load cart: %w", err)
	}

	// 3. Validate cart is not empty
	cartItems := cart.GetItems()
	if len(cartItems) == 0 {
		return nil, fmt.Errorf("cannot create order from empty cart")
	}

	// 4. Validate delivery info if provided
	if deliveryInfo != nil && !deliveryInfo.IsValid() {
		return nil, fmt.Errorf("invalid delivery info")
	}

	// 5. Convert cart items to order items
	orderItems := convertCartToOrderItems(cartItems)

	// 6. Create order from cart items
	order := orderDomain.NewOrderState(customerID)
	if err := order.CreateOrder(orderItems); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 7. Set delivery info if provided
	if deliveryInfo != nil {
		order.SetDeliveryInfo(*deliveryInfo)
	}

	// 8. Clear cart
	cart.Reset()

	// 9. Save order (uses tx from ctx)
	if err := uc.orderRepo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// 10. Save cart (uses tx from ctx)
	if err := uc.cartRepo.Save(ctx, cart); err != nil {
		return nil, fmt.Errorf("failed to save cart: %w", err)
	}

	// 11. Commit transaction
	if err := uc.uow.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &CreateOrderFromCartResult{
		Order: order,
	}, nil
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
