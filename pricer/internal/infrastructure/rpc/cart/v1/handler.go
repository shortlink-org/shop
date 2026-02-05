package v1

import (
	"context"
	"fmt"

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
	if req == nil || req.GetCart() == nil {
		return &CalculateTotalResponse{}, nil
	}

	cart, err := protoToDomainCart(req.GetCart())
	if err != nil {
		return nil, fmt.Errorf("invalid cart request: %w", err)
	}

	discountParams := stringMapToInterface(req.GetDiscountParams())
	taxParams := stringMapToInterface(req.GetTaxParams())

	cmd := calculate_total.NewCommand(cart, discountParams, taxParams)

	total, err := h.calculateTotalHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("calculate total: %w", err)
	}

	return &CalculateTotalResponse{
		Total: domainToProtoCartTotal(&total),
	}, nil
}

func protoToDomainCart(p *Cart) (*domain.Cart, error) {
	if p == nil {
		return nil, nil
	}

	items := make([]domain.CartItem, 0, len(p.GetItems()))
	for _, it := range p.GetItems() {
		goodID, err := uuid.Parse(it.GetProductId())
		if err != nil {
			return nil, fmt.Errorf("item product_id: %w", err)
		}

		price, err := decimal.NewFromString(it.GetPrice())
		if err != nil {
			return nil, fmt.Errorf("item price: %w", err)
		}

		items = append(items, domain.CartItem{
			GoodID:   goodID,
			Quantity: it.GetQuantity(),
			Price:    price,
		})
	}

	customerID, err := uuid.Parse(p.GetCustomerId())
	if err != nil {
		return nil, fmt.Errorf("customer_id: %w", err)
	}

	return &domain.Cart{
		Items:      items,
		CustomerID: customerID,
	}, nil
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

func stringMapToInterface(m map[string]string) map[string]any {
	if m == nil {
		return nil
	}

	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}

	return result
}
