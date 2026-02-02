package domain

import "github.com/shopspring/decimal"

type CartTotal struct {
	TotalTax      decimal.Decimal `json:"totalTax"`
	TotalDiscount decimal.Decimal `json:"totalDiscount"`
	FinalPrice    decimal.Decimal `json:"finalPrice"`
	Policies      []string        `json:"policies"`
}
