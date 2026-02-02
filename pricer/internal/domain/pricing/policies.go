package pricing

import (
	"context"

	"github.com/shortlink-org/shop/pricer/internal/domain"
	"github.com/shortlink-org/shop/pricer/internal/infrastructure/policy_evaluator"
)

// DiscountPolicy wraps a policy evaluator for discounts.
type DiscountPolicy struct {
	Evaluator policy_evaluator.PolicyEvaluator
}

// Evaluate evaluates the discount policy.
func (p *DiscountPolicy) Evaluate(ctx context.Context, cart *domain.Cart, params map[string]interface{}) (float64, error) {
	return p.Evaluator.Evaluate(ctx, cart, params)
}

// TaxPolicy wraps a policy evaluator for taxes.
type TaxPolicy struct {
	Evaluator policy_evaluator.PolicyEvaluator
}

// Evaluate evaluates the tax policy.
func (p *TaxPolicy) Evaluate(ctx context.Context, cart *domain.Cart, params map[string]interface{}) (float64, error) {
	return p.Evaluator.Evaluate(ctx, cart, params)
}
