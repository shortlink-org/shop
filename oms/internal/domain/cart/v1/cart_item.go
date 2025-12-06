package v1

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	pricing "github.com/shortlink-org/shop/oms/internal/domain/pricing"
)

// CartItem validation errors
var (
	ErrCartItemGoodIdZero           = errors.New("cart item good ID cannot be zero")
	ErrCartItemQuantityZero         = errors.New("cart item quantity must be greater than zero")
	ErrCartItemPriceNegative        = errors.New("cart item price cannot be negative")
	ErrCartItemDiscountNegative     = errors.New("cart item discount cannot be negative")
	ErrCartItemTaxNegative          = errors.New("cart item tax cannot be negative")
	ErrCartItemDiscountExceedsPrice = errors.New("cart item discount cannot exceed price")
)

// CartItem represents an immutable cart item value object.
// All fields are private and can only be set through constructors.
type CartItem struct {
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

// NewCartItem creates a new CartItem with required fields only.
// Price, discount, and tax are set to zero and should be set using NewCartItemWithPricing.
func NewCartItem(goodId uuid.UUID, quantity int32) (CartItem, error) {
	if goodId == uuid.Nil {
		return CartItem{}, ErrCartItemGoodIdZero
	}
	if quantity <= 0 {
		return CartItem{}, ErrCartItemQuantityZero
	}

	return CartItem{
		goodId:   goodId,
		quantity: quantity,
		price:    decimal.Zero,
		discount: decimal.Zero,
		tax:      decimal.Zero,
	}, nil
}

// NewCartItemWithPricing creates a new CartItem with all pricing information.
// This constructor validates all pricing fields according to business rules.
func NewCartItemWithPricing(
	goodId uuid.UUID,
	quantity int32,
	price decimal.Decimal,
	discount decimal.Decimal,
	tax decimal.Decimal,
) (CartItem, error) {
	if goodId == uuid.Nil {
		return CartItem{}, ErrCartItemGoodIdZero
	}
	if quantity <= 0 {
		return CartItem{}, ErrCartItemQuantityZero
	}
	if price.IsNegative() {
		return CartItem{}, ErrCartItemPriceNegative
	}
	if discount.IsNegative() {
		return CartItem{}, ErrCartItemDiscountNegative
	}
	if tax.IsNegative() {
		return CartItem{}, ErrCartItemTaxNegative
	}
	if discount.GreaterThan(price) {
		return CartItem{}, fmt.Errorf("%w: discount %s exceeds price %s", ErrCartItemDiscountExceedsPrice, discount.String(), price.String())
	}

	return CartItem{
		goodId:   goodId,
		quantity: quantity,
		price:    price,
		discount: discount,
		tax:      tax,
	}, nil
}

// WithPricing returns a new CartItem with updated pricing information.
// This preserves immutability by creating a new instance.
func (c CartItem) WithPricing(price, discount, tax decimal.Decimal) (CartItem, error) {
	return NewCartItemWithPricing(c.goodId, c.quantity, price, discount, tax)
}

// WithQuantity returns a new CartItem with updated quantity.
// This preserves immutability by creating a new instance.
func (c CartItem) WithQuantity(quantity int32) (CartItem, error) {
	if quantity <= 0 {
		return CartItem{}, ErrCartItemQuantityZero
	}

	return CartItem{
		goodId:   c.goodId,
		quantity: quantity,
		price:    c.price,
		discount: c.discount,
		tax:      c.tax,
	}, nil
}

// WithPricePolicy applies the provided PricePolicy to the item and returns a new priced item.
func (c CartItem) WithPricePolicy(policy pricing.PricePolicy) (CartItem, error) {
	if policy == nil {
		policy = pricing.NoopPricePolicy{}
	}

	quote, err := policy.Quote(c.goodId, c.quantity)
	if err != nil {
		return CartItem{}, err
	}

	return NewCartItemWithPricing(c.goodId, c.quantity, quote.UnitPrice, quote.Discount, quote.Tax)
}

// GetGoodId returns the good ID.
func (c CartItem) GetGoodId() uuid.UUID {
	return c.goodId
}

// GetQuantity returns the quantity.
func (c CartItem) GetQuantity() int32 {
	return c.quantity
}

// GetPrice returns the price per unit.
func (c CartItem) GetPrice() decimal.Decimal {
	return c.price
}

// GetDiscount returns the discount per unit.
func (c CartItem) GetDiscount() decimal.Decimal {
	return c.discount
}

// GetTax returns the tax per unit.
func (c CartItem) GetTax() decimal.Decimal {
	return c.tax
}

// GetPriceAfterDiscount returns the price after discount (price - discount).
func (c CartItem) GetPriceAfterDiscount() decimal.Decimal {
	priceAfterDiscount := c.price.Sub(c.discount)
	if priceAfterDiscount.IsNegative() {
		return decimal.Zero
	}
	return priceAfterDiscount
}

// GetSubtotal returns the subtotal for this item (price after discount * quantity).
func (c CartItem) GetSubtotal() decimal.Decimal {
	return c.GetPriceAfterDiscount().Mul(decimal.NewFromInt32(c.quantity))
}

// GetTotalTax returns the total tax for this item (tax per unit * quantity).
func (c CartItem) GetTotalTax() decimal.Decimal {
	return c.tax.Mul(decimal.NewFromInt32(c.quantity))
}

// GetTotal returns the total for this item (subtotal + total tax).
func (c CartItem) GetTotal() decimal.Decimal {
	return c.GetSubtotal().Add(c.GetTotalTax())
}

// IsValid checks if the cart item is valid.
func (c CartItem) IsValid() bool {
	return c.goodId != uuid.Nil &&
		c.quantity > 0 &&
		!c.price.IsNegative() &&
		!c.discount.IsNegative() &&
		!c.tax.IsNegative() &&
		!c.discount.GreaterThan(c.price)
}
