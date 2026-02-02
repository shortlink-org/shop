package calculate_total

import (
	"github.com/shortlink-org/shop/pricer/internal/domain"
)

// Command represents a command to calculate cart totals (discount + tax).
type Command struct {
	Cart           *domain.Cart
	DiscountParams map[string]interface{}
	TaxParams      map[string]interface{}
}

// NewCommand creates a new CalculateTotal command.
func NewCommand(cart *domain.Cart, discountParams, taxParams map[string]interface{}) Command {
	return Command{
		Cart:           cart,
		DiscountParams: discountParams,
		TaxParams:      taxParams,
	}
}
