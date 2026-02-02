package calculate_total

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shopspring/decimal"

	logger "github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/shop/pricer/internal/domain"
	"github.com/shortlink-org/shop/pricer/internal/domain/ports"
	"github.com/shortlink-org/shop/pricer/internal/domain/pricing"
)

// Handler handles CalculateTotal commands.
type Handler struct {
	log            logger.Logger
	discountPolicy *pricing.DiscountPolicy
	taxPolicy      *pricing.TaxPolicy
	policyNames    []string
}

// NewHandler creates a new CalculateTotal handler.
func NewHandler(
	log logger.Logger,
	discountPolicy *pricing.DiscountPolicy,
	taxPolicy *pricing.TaxPolicy,
	policyNames []string,
) *Handler {
	return &Handler{
		log:            log,
		discountPolicy: discountPolicy,
		taxPolicy:      taxPolicy,
		policyNames:    policyNames,
	}
}

// Handle executes the CalculateTotal command.
func (h *Handler) Handle(ctx context.Context, cmd Command) (domain.CartTotal, error) {
	var total domain.CartTotal

	if cmd.Cart == nil {
		return total, nil
	}

	// Evaluate Discount Policy
	h.log.InfoWithContext(ctx, "Evaluating discount policy", slog.Any("customer_id", cmd.Cart.CustomerID))
	totalDiscountFloat, err := h.discountPolicy.Evaluate(ctx, cmd.Cart, cmd.DiscountParams)
	if err != nil {
		return total, fmt.Errorf("failed to evaluate discount policy: %w", err)
	}

	h.log.InfoWithContext(ctx, "Discount calculated", slog.Float64("total_discount", totalDiscountFloat))
	totalDiscount := decimal.NewFromFloat(totalDiscountFloat)

	// Evaluate Tax Policy
	h.log.InfoWithContext(ctx, "Evaluating tax policy", slog.Any("customer_id", cmd.Cart.CustomerID))
	totalTaxFloat, err := h.taxPolicy.Evaluate(ctx, cmd.Cart, cmd.TaxParams)
	if err != nil {
		return total, fmt.Errorf("failed to evaluate tax policy: %w", err)
	}

	h.log.InfoWithContext(ctx, "Tax calculated", slog.Float64("total_tax", totalTaxFloat))
	totalTax := decimal.NewFromFloat(totalTaxFloat)

	// Calculate subtotal
	h.log.InfoWithContext(ctx, "Calculating final price", slog.Any("customer_id", cmd.Cart.CustomerID))
	subtotal := decimal.Zero
	for _, item := range cmd.Cart.Items {
		itemSubtotal := item.Price.Mul(decimal.NewFromInt32(item.Quantity))
		subtotal = subtotal.Add(itemSubtotal)
		h.log.InfoWithContext(ctx, "Item subtotal calculated",
			slog.Any("item_id", item.GoodID),
			slog.String("item_subtotal", itemSubtotal.StringFixed(2)),
		)
	}

	// Cap discount at subtotal to avoid negative final price
	if totalDiscount.GreaterThan(subtotal) {
		totalDiscount = subtotal
	}
	finalPrice := subtotal.Sub(totalDiscount).Add(totalTax)

	total = domain.CartTotal{
		TotalTax:      totalTax,
		TotalDiscount: totalDiscount,
		FinalPrice:    finalPrice,
		Policies:      h.policyNames,
	}

	h.log.InfoWithContext(ctx, "Final price calculated",
		slog.Any("customer_id", cmd.Cart.CustomerID),
		slog.String("final_price", finalPrice.StringFixed(2)),
	)

	return total, nil
}

// Ensure Handler implements CommandHandlerWithResult interface.
var _ ports.CommandHandlerWithResult[Command, domain.CartTotal] = (*Handler)(nil)
