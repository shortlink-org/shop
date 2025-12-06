package v1

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Order validation errors
var (
	ErrOrderItemsEmpty        = errors.New("order must have at least one item")
	ErrOrderItemInvalid       = errors.New("order item is invalid")
	ErrOrderItemQuantityZero  = errors.New("order item quantity must be greater than zero")
	ErrOrderItemPriceNegative = errors.New("order item price cannot be negative")
	ErrOrderItemPriceZero     = errors.New("order item price must be greater than zero")
	ErrOrderTotalWeightExceeded = errors.New("total order weight exceeds maximum allowed")
	ErrOrderTotalItemsExceeded  = errors.New("total order items count exceeds maximum allowed")
	ErrOrderItemsDuplicate      = errors.New("order contains duplicate items")
	ErrOrderInvalidStateTransition = errors.New("invalid state transition for order")
)

// Order invariants constants
const (
	// MaxOrderWeightKg is the maximum total weight allowed for an order in kilograms
	MaxOrderWeightKg = 500.0
	// MaxOrderItems is the maximum number of distinct items allowed in an order
	MaxOrderItems = 100
	// MinOrderItems is the minimum number of items required in an order
	MinOrderItems = 1
)

// ValidateOrderItems validates that order items meet all business rules and invariants.
func ValidateOrderItems(items Items) error {
	if len(items) == 0 {
		return ErrOrderItemsEmpty
	}

	if len(items) > MaxOrderItems {
		return fmt.Errorf("%w: %d items, maximum is %d", ErrOrderTotalItemsExceeded, len(items), MaxOrderItems)
	}

	// Track seen good IDs to detect duplicates
	seenGoodIds := make(map[string]bool)
	totalWeight := 0.0

	for i, item := range items {
		// Validate individual item
		if err := ValidateOrderItem(item); err != nil {
			return fmt.Errorf("item %d: %w", i, err)
		}

		// Check for duplicate items
		goodIdStr := item.GetGoodId().String()
		if seenGoodIds[goodIdStr] {
			return fmt.Errorf("%w: good ID %s appears multiple times", ErrOrderItemsDuplicate, goodIdStr)
		}
		seenGoodIds[goodIdStr] = true

		// Note: We don't have weight per item in the current Item model
		// This would need to be added if we want to validate total weight
		// For now, we validate structure only
		_ = totalWeight // placeholder for future weight validation
	}

	return nil
}

// ValidateOrderItem validates a single order item.
func ValidateOrderItem(item Item) error {
	if item.GetGoodId() == uuid.Nil {
		return fmt.Errorf("%w: good ID is zero", ErrOrderItemInvalid)
	}

	if item.GetQuantity() <= 0 {
		return ErrOrderItemQuantityZero
	}

	price := item.GetPrice()
	if price.IsNegative() {
		return ErrOrderItemPriceNegative
	}

	if price.IsZero() {
		return ErrOrderItemPriceZero
	}

	return nil
}

// ValidateOrderStateTransition validates that a state transition is allowed given the current order state.
func ValidateOrderStateTransition(currentStatus OrderStatus, targetStatus OrderStatus, items Items) error {
	// Validate items are not empty when transitioning to PROCESSING or COMPLETED
	if targetStatus == OrderStatus_ORDER_STATUS_PROCESSING || targetStatus == OrderStatus_ORDER_STATUS_COMPLETED {
		if len(items) == 0 {
			return fmt.Errorf("%w: cannot transition to %s with empty items", ErrOrderInvalidStateTransition, targetStatus)
		}

		// Validate items meet business rules
		if err := ValidateOrderItems(items); err != nil {
			return fmt.Errorf("%w: %w", ErrOrderInvalidStateTransition, err)
		}
	}

	// Additional state-specific validations can be added here

	return nil
}

// CalculateTotalWeight calculates the total weight of order items.
// Note: Current Item model doesn't include weight, this is a placeholder for future enhancement.
func CalculateTotalWeight(items Items) float64 {
	// TODO: Add weight field to Item or fetch from product catalog
	// For now, return 0 as placeholder
	totalWeight := 0.0
	for range items {
		// Would need: totalWeight += item.GetWeight() * float64(item.GetQuantity())
	}
	return totalWeight
}

// CalculateTotalPrice calculates the total price of all items in the order.
func CalculateTotalPrice(items Items) decimal.Decimal {
	total := decimal.Zero
	for _, item := range items {
		itemTotal := item.GetPrice().Mul(decimal.NewFromInt32(item.GetQuantity()))
		total = total.Add(itemTotal)
	}
	return total
}

// ValidateOrderTotalWeight validates that the total order weight doesn't exceed maximum.
func ValidateOrderTotalWeight(totalWeightKg float64) error {
	if totalWeightKg > MaxOrderWeightKg {
		return fmt.Errorf("%w: %.2f kg, maximum is %.2f kg", ErrOrderTotalWeightExceeded, totalWeightKg, MaxOrderWeightKg)
	}
	return nil
}

