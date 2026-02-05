package calculate_total

import (
	"github.com/shortlink-org/shop/pricer/internal/domain"
)

// Command represents a command to calculate cart totals (discount + tax).
type Command struct {
	Cart           *domain.Cart
	DiscountParams map[string]any
	TaxParams      map[string]any
}

// NewCommand creates a new CalculateTotal command.
func NewCommand(cart *domain.Cart, discountParams, taxParams map[string]any) Command {
	return Command{
		Cart:           cart,
		DiscountParams: discountParams,
		TaxParams:      taxParams,
	}
}
