package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PricerClient is the interface for the pricing service client.
type PricerClient interface {
	// CalculateTotal calculates the total price, tax, and discounts for a cart.
	CalculateTotal(ctx context.Context, req CalculateTotalRequest) (*CalculateTotalResponse, error)
}

// CalculateTotalRequest is the request for calculating cart totals.
type CalculateTotalRequest struct {
	Cart           CartData
	DiscountParams map[string]string
	TaxParams      map[string]string
}

// CalculateTotalResponse is the response after calculating totals.
type CalculateTotalResponse struct {
	TotalTax      decimal.Decimal
	TotalDiscount decimal.Decimal
	FinalPrice    decimal.Decimal
	Subtotal      decimal.Decimal
	Policies      []string
}

// CartData represents cart data for pricing calculation.
type CartData struct {
	CustomerID uuid.UUID
	Items      []CartItemData
}

// CartItemData represents a cart item for pricing calculation.
type CartItemData struct {
	ProductID uuid.UUID       // Good/product identifier
	Quantity  int32           // Number of units
	UnitPrice decimal.Decimal // Price per unit (before discount/tax)
}
