package pricing

import (
	"context"
	"fmt"

	"github.com/shortlink-org/shop/pricer/internal/domain"
	"github.com/shortlink-org/shop/pricer/internal/infrastructure/policy_evaluator"
)

// DiscountPolicy wraps a policy evaluator for discounts.
type DiscountPolicy struct {
	Evaluator policy_evaluator.PolicyEvaluator
}

// Evaluate evaluates the discount policy.
func (p *DiscountPolicy) Evaluate(ctx context.Context, cart *domain.Cart, params map[string]any) (float64, error) {
	v, err := p.Evaluator.Evaluate(ctx, cart, params)
	if err != nil {
		return 0, fmt.Errorf("discount policy: %w", err)
	}

	return v, nil
}

// TaxPolicy wraps a policy evaluator for taxes.
type TaxPolicy struct {
	Evaluator policy_evaluator.PolicyEvaluator
}

// Evaluate evaluates the tax policy.
func (p *TaxPolicy) Evaluate(ctx context.Context, cart *domain.Cart, params map[string]any) (float64, error) {
	v, err := p.Evaluator.Evaluate(ctx, cart, params)
	if err != nil {
		return 0, fmt.Errorf("tax policy: %w", err)
	}

	return v, nil
}
