package v1

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/shortlink-org/shop/pricer/internal/domain"
	"github.com/shortlink-org/shop/pricer/internal/usecases/cart/command/calculate_total"
)

// CartHandler implements CartServiceServer
type CartHandler struct {
	UnimplementedCartServiceServer
	calculateTotalHandler *calculate_total.Handler
}

// NewCartHandler creates a new CartHandler
func NewCartHandler(calculateTotalHandler *calculate_total.Handler) *CartHandler {
	return &CartHandler{
		calculateTotalHandler: calculateTotalHandler,
	}
}

// CalculateTotal calculates the total price, tax, and discounts for a cart
func (h *CartHandler) CalculateTotal(ctx context.Context, req *CalculateTotalRequest) (*CalculateTotalResponse, error) {
	if req == nil || req.Cart == nil {
		return &CalculateTotalResponse{}, nil
	}

	cart := protoToDomainCart(req.Cart)
	discountParams := stringMapToInterface(req.GetDiscountParams())
	taxParams := stringMapToInterface(req.GetTaxParams())

	cmd := calculate_total.NewCommand(cart, discountParams, taxParams)
	total, err := h.calculateTotalHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &CalculateTotalResponse{
		Total: domainToProtoCartTotal(&total),
	}, nil
}

func protoToDomainCart(p *Cart) *domain.Cart {
	if p == nil {
		return nil
	}

	items := make([]domain.CartItem, 0, len(p.GetItems()))
	for _, it := range p.GetItems() {
		goodID, _ := uuid.Parse(it.GetProductId())
		price, _ := decimal.NewFromString(it.GetPrice())
		items = append(items, domain.CartItem{
			GoodID:   goodID,
			Quantity: it.GetQuantity(),
			Price:    price,
		})
	}

	customerID, _ := uuid.Parse(p.GetCustomerId())

	return &domain.Cart{
		Items:      items,
		CustomerID: customerID,
	}
}

func domainToProtoCartTotal(t *domain.CartTotal) *CartTotal {
	if t == nil {
		return nil
	}

	return &CartTotal{
		TotalTax:      t.TotalTax.String(),
		TotalDiscount: t.TotalDiscount.String(),
		FinalPrice:    t.FinalPrice.String(),
		Policies:      t.Policies,
	}
}

func stringMapToInterface(m map[string]string) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
