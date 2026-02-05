package pricing

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Quote describes the pricing components for a single item.
type Quote struct {
	// UnitPrice is the base price per unit before discounts and taxes.
	UnitPrice decimal.Decimal
	// Discount is the per-unit discount amount.
	Discount decimal.Decimal
	// Tax is the per-unit tax amount.
	Tax decimal.Decimal
}

// FinalUnitPrice returns the per-unit amount after discount and tax.
func (q Quote) FinalUnitPrice() decimal.Decimal {
	return q.UnitPrice.Sub(q.Discount).Add(q.Tax)
}

// PricePolicy is responsible for providing pricing information for a good.
type PricePolicy interface {
	Quote(goodID uuid.UUID, quantity int32) (Quote, error)
}

// NoopPricePolicy returns zeroed quotes until the real pricer is integrated.
type NoopPricePolicy struct{}

// Quote implements PricePolicy.
func (NoopPricePolicy) Quote(goodID uuid.UUID, quantity int32) (Quote, error) {
	return Quote{}, nil
}
