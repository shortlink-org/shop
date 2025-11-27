package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shopspring/decimal"

	logger "github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/shop/pricer/internal/domain"
)

// NewCartService creates a new CartService
func NewCartService(log logger.Logger, discountPolicy DiscountPolicy, taxPolicy TaxPolicy, policyNames []string) *CartService {
	return &CartService{
		log: log,

		DiscountPolicy: discountPolicy,
		TaxPolicy:      taxPolicy,
		PolicyNames:    policyNames,
	}
}

// CalculateTotal computes the total price, applying discounts and taxes
func (s *CartService) CalculateTotal(ctx context.Context, cart *domain.Cart, discountParams, taxParams map[string]interface{}) (CartTotal, error) {
	var total CartTotal

	// Evaluate Discount Policy
	s.log.InfoWithContext(ctx, "Evaluating discount policy", slog.Any("customer_id", cart.CustomerID))
	totalDiscountFloat, err := s.DiscountPolicy.Evaluate(ctx, cart, discountParams)
	if err != nil {
		return total, fmt.Errorf("failed to evaluate discount policy: %w", err)
	}

	s.log.InfoWithContext(ctx, "Discount calculated", slog.Float64("total_discount", totalDiscountFloat))
	totalDiscount := decimal.NewFromFloat(totalDiscountFloat)

	// Evaluate Tax Policy
	s.log.InfoWithContext(ctx, "Evaluating tax policy", slog.Any("customer_id", cart.CustomerID))
	totalTaxFloat, err := s.TaxPolicy.Evaluate(ctx, cart, taxParams)
	if err != nil {
		return total, fmt.Errorf("failed to evaluate tax policy: %w", err)
	}

	s.log.InfoWithContext(ctx, "Tax calculated", slog.Float64("total_tax", totalTaxFloat))
	totalTax := decimal.NewFromFloat(totalTaxFloat)

	// Calculate Final Price
	s.log.InfoWithContext(ctx, "Calculating final price", slog.Any("customer_id", cart.CustomerID))
	finalPrice := decimal.Zero
	for _, item := range cart.Items {
		// Calculate per-item total: (Price + Tax - Discount) * Quantity
		itemTotal := item.Price.Add(totalTax).Sub(totalDiscount).Mul(decimal.NewFromInt32(item.Quantity))
		finalPrice = finalPrice.Add(itemTotal)
		s.log.InfoWithContext(ctx, "Item total calculated",
			slog.Any("item_id", item.GoodID),
			slog.String("item_total", itemTotal.StringFixed(2)),
		)
	}

	// Prepare the CartTotal
	total = CartTotal{
		TotalTax:      totalTax,
		TotalDiscount: totalDiscount,
		FinalPrice:    finalPrice,
		Policies:      s.PolicyNames,
	}

	s.log.InfoWithContext(ctx, "Final price calculated",
		slog.Any("customer_id", cart.CustomerID),
		slog.String("final_price", finalPrice.StringFixed(2)),
	)

	return total, nil
}
