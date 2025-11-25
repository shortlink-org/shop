package application

import (
	"github.com/shopspring/decimal"

	logger "github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/shop/pricer/internal/infrastructure/policy_evaluator"
)

// CartTotal represents the total calculation result
type CartTotal struct {
	TotalTax      decimal.Decimal `json:"totalTax"`
	TotalDiscount decimal.Decimal `json:"totalDiscount"`
	FinalPrice    decimal.Decimal `json:"finalPrice"`
	Policies      []string        `json:"policies"`
}

// CartService orchestrates cart operations
type CartService struct {
	log logger.Logger

	DiscountPolicy policy_evaluator.PolicyEvaluator
	TaxPolicy      policy_evaluator.PolicyEvaluator
	PolicyNames    []string
}

type DiscountPolicy policy_evaluator.PolicyEvaluator
type TaxPolicy policy_evaluator.PolicyEvaluator
