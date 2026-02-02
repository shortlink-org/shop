package domain

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CartItem struct {
	GoodID   uuid.UUID       `json:"productId"`
	Quantity int32           `json:"quantity"`
	Price    decimal.Decimal `json:"price"`
}

type Cart struct {
	Items      []CartItem `json:"items"`
	CustomerID uuid.UUID  `json:"customerId"`
}

func (c *Cart) AddItem(item CartItem) {
	c.Items = append(c.Items, item)
}
