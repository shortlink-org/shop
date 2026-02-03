package create_order_from_cart

import (
	cartItemsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/items/v1"
	orderDomain "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// cartItemsToLines maps cart items to neutral Line slice for order domain.
func cartItemsToLines(cartItems cartItemsv1.Items) []orderDomain.Line {
	lines := make([]orderDomain.Line, 0, len(cartItems))
	for _, item := range cartItems {
		lines = append(lines, orderDomain.Line{
			ProductID: item.GetGoodId(),
			Qty:       item.GetQuantity(),
			UnitPrice: item.GetPrice(),
		})
	}
	return lines
}
