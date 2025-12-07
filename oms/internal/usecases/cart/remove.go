package cart

import (
	"context"
	"fmt"

	domain "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	v1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	"github.com/shortlink-org/shop/oms/internal/usecases/cart/dto"
)

// Remove removes items from the cart.
func (uc *UC) Remove(ctx context.Context, in *domain.State) error {
	workflowId := fmt.Sprintf("cart-%s", in.GetCustomerId().String())

	request := dto.CartStateToCartEvent(in, v1.Event_EVENT_REMOVE)
	err := uc.temporalClient.SignalWorkflow(ctx, workflowId, "", v1.Event_EVENT_REMOVE.String(), request)
	if err != nil {
		return err
	}

	// Update index: remove goods from customer's cart index
	for _, item := range in.GetItems() {
		uc.goodsIndex.RemoveGoodFromCart(item.GetGoodId(), in.GetCustomerId())
	}

	return nil
}
