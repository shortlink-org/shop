package v1

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CartItem represents a cart item.
type CartItem struct {
	// goodId is the good ID
	goodId uuid.UUID
	// quantity is the quantity of the good
	quantity int32
	// price is the price of the good
	price decimal.Decimal
	// discount is the discount of the good
	discount decimal.Decimal
	// tax is the tax of the good
	tax decimal.Decimal
}

// NewCartItem creates a new CartItem.
func NewCartItem(goodId uuid.UUID, quantity int32) CartItem {
	return CartItem{
		goodId:   goodId,
		quantity: quantity,
	}
}

// GetGoodId returns the good ID.
func (c CartItem) GetGoodId() uuid.UUID {
	return c.goodId
}

// GetQuantity returns the quantity.
func (c CartItem) GetQuantity() int32 {
	return c.quantity
}

// GetPrice returns the price.
func (c CartItem) GetPrice() decimal.Decimal {
	return c.price
}
