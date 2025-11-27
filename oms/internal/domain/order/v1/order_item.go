package v1

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Items represent a list of order items.
type Items []Item

// Item represents an item in the order.
type Item struct {
	goodId   uuid.UUID
	quantity int32
	price    decimal.Decimal
}

// NewItem creates a new item.
func NewItem(goodId uuid.UUID, quantity int32, price decimal.Decimal) Item {
	return Item{
		goodId:   goodId,
		quantity: quantity,
		price:    price,
	}
}

// GetGoodId returns the value of the goodId field.
func (m Item) GetGoodId() uuid.UUID {
	return m.goodId
}

// GetQuantity returns the value of the quantity field.
func (m Item) GetQuantity() int32 {
	return m.quantity
}

// GetPrice returns the value of the price field.
func (m Item) GetPrice() decimal.Decimal {
	return m.price
}
