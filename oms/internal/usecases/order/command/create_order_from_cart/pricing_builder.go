package create_order_from_cart

import (
	"github.com/google/uuid"

	cartItemsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/items/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// PricerRequestBuilder builds a CalculateTotalRequest with optional discount/tax params.
type PricerRequestBuilder struct {
	req ports.CalculateTotalRequest
}

// NewPricerRequestBuilder starts a new builder from customer ID and cart items.
func NewPricerRequestBuilder(customerID uuid.UUID, items cartItemsv1.Items) *PricerRequestBuilder {
	cartItems := make([]ports.CartItemData, 0, len(items))
	for _, item := range items {
		cartItems = append(cartItems, ports.CartItemData{
			ProductID: item.GetGoodId(),
			Quantity:  item.GetQuantity(),
			UnitPrice: item.GetPrice(),
		})
	}
	return &PricerRequestBuilder{
		req: ports.CalculateTotalRequest{
			Cart: ports.CartData{
				CustomerID: customerID,
				Items:      cartItems,
			},
			DiscountParams: nil,
			TaxParams:      nil,
		},
	}
}

// WithDiscountParam adds a discount parameter (e.g. promo code, customer segment).
func (b *PricerRequestBuilder) WithDiscountParam(k, v string) *PricerRequestBuilder {
	if b.req.DiscountParams == nil {
		b.req.DiscountParams = make(map[string]string)
	}
	b.req.DiscountParams[k] = v
	return b
}

// WithTaxParam adds a tax parameter (e.g. country, city, postalCode).
func (b *PricerRequestBuilder) WithTaxParam(k, v string) *PricerRequestBuilder {
	if b.req.TaxParams == nil {
		b.req.TaxParams = make(map[string]string)
	}
	b.req.TaxParams[k] = v
	return b
}

// Build returns the built request.
func (b *PricerRequestBuilder) Build() ports.CalculateTotalRequest {
	return b.req
}
