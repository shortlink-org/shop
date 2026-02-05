package v1

import (
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	pricing "github.com/shortlink-org/shop/oms/internal/domain/pricing"
)

// Item validation errors
var (
	ErrItemGoodIdZero           = errors.New("item good ID cannot be zero")
	ErrItemQuantityZero         = errors.New("item quantity must be greater than zero")
	ErrItemPriceNegative        = errors.New("item price cannot be negative")
	ErrItemDiscountNegative     = errors.New("item discount cannot be negative")
	ErrItemTaxNegative          = errors.New("item tax cannot be negative")
	ErrItemDiscountExceedsPrice = errors.New("item discount cannot exceed price")
)

// Item represents an immutable cart item.
// All fields are private and can only be set through constructors.
type Item struct {
	// goodId is the good ID
	goodId uuid.UUID
	// quantity is the quantity of the good
	quantity int32
	// price is the price per unit of the good
	price decimal.Decimal
	// discount is the discount amount per unit
	discount decimal.Decimal
	// tax is the tax amount per unit
	tax decimal.Decimal
}

// NewItem creates a new Item with required fields only.
// Price, discount, and tax are set to zero and should be set using NewItemWithPricing.
// Validation is performed using the Specification pattern from rules package.
func NewItem(goodId uuid.UUID, quantity int32) (Item, error) {
	item := Item{
		goodId:   goodId,
		quantity: quantity,
		price:    decimal.Zero,
		discount: decimal.Zero,
		tax:      decimal.Zero,
	}

	// Import cycle prevention: validate inline for basic fields
	if goodId == uuid.Nil {
		return Item{}, ErrItemGoodIdZero
	}

	if quantity <= 0 {
		return Item{}, ErrItemQuantityZero
	}

	return item, nil
}

// NewItemWithPricing creates a new Item with all pricing information.
// Validation is performed using the Specification pattern from rules package.
func NewItemWithPricing(
	goodId uuid.UUID,
	quantity int32,
	price decimal.Decimal,
	discount decimal.Decimal,
	tax decimal.Decimal,
) (Item, error) {
	item := Item{
		goodId:   goodId,
		quantity: quantity,
		price:    price,
		discount: discount,
		tax:      tax,
	}

	// Import cycle prevention: validate inline
	if goodId == uuid.Nil {
		return Item{}, ErrItemGoodIdZero
	}

	if quantity <= 0 {
		return Item{}, ErrItemQuantityZero
	}

	if price.IsNegative() {
		return Item{}, ErrItemPriceNegative
	}

	if discount.IsNegative() {
		return Item{}, ErrItemDiscountNegative
	}

	if tax.IsNegative() {
		return Item{}, ErrItemTaxNegative
	}

	if discount.GreaterThan(price) {
		return Item{}, ErrItemDiscountExceedsPrice
	}

	return item, nil
}

// WithPricing returns a new Item with updated pricing information.
// This preserves immutability by creating a new instance.
func (i Item) WithPricing(price, discount, tax decimal.Decimal) (Item, error) {
	return NewItemWithPricing(i.goodId, i.quantity, price, discount, tax)
}

// WithQuantity returns a new Item with updated quantity.
// This preserves immutability by creating a new instance.
func (i Item) WithQuantity(quantity int32) (Item, error) {
	if quantity <= 0 {
		return Item{}, ErrItemQuantityZero
	}

	return Item{
		goodId:   i.goodId,
		quantity: quantity,
		price:    i.price,
		discount: i.discount,
		tax:      i.tax,
	}, nil
}

// WithPricePolicy applies the provided PricePolicy to the item and returns a new priced item.
func (i Item) WithPricePolicy(policy pricing.PricePolicy) (Item, error) {
	if policy == nil {
		policy = pricing.NoopPricePolicy{}
	}

	quote, err := policy.Quote(i.goodId, i.quantity)
	if err != nil {
		return Item{}, err
	}

	return NewItemWithPricing(i.goodId, i.quantity, quote.UnitPrice, quote.Discount, quote.Tax)
}

// GetGoodId returns the good ID.
func (i Item) GetGoodId() uuid.UUID {
	return i.goodId
}

// GetQuantity returns the quantity.
func (i Item) GetQuantity() int32 {
	return i.quantity
}

// GetPrice returns the price per unit.
func (i Item) GetPrice() decimal.Decimal {
	return i.price
}

// GetDiscount returns the discount per unit.
func (i Item) GetDiscount() decimal.Decimal {
	return i.discount
}

// GetTax returns the tax per unit.
func (i Item) GetTax() decimal.Decimal {
	return i.tax
}

// GetPriceAfterDiscount returns the price after discount (price - discount).
func (i Item) GetPriceAfterDiscount() decimal.Decimal {
	priceAfterDiscount := i.price.Sub(i.discount)
	if priceAfterDiscount.IsNegative() {
		return decimal.Zero
	}

	return priceAfterDiscount
}

// GetSubtotal returns the subtotal for this item (price after discount * quantity).
func (i Item) GetSubtotal() decimal.Decimal {
	return i.GetPriceAfterDiscount().Mul(decimal.NewFromInt32(i.quantity))
}

// GetTotalTax returns the total tax for this item (tax per unit * quantity).
func (i Item) GetTotalTax() decimal.Decimal {
	return i.tax.Mul(decimal.NewFromInt32(i.quantity))
}

// GetTotal returns the total for this item (subtotal + total tax).
func (i Item) GetTotal() decimal.Decimal {
	return i.GetSubtotal().Add(i.GetTotalTax())
}

// IsValid checks if the item is valid.
func (i Item) IsValid() bool {
	return i.goodId != uuid.Nil &&
		i.quantity > 0 &&
		!i.price.IsNegative() &&
		!i.discount.IsNegative() &&
		!i.tax.IsNegative() &&
		!i.discount.GreaterThan(i.price)
}
