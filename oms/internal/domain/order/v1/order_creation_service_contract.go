package v1

import "fmt"

// StockCheckerContract defines the contract for stock availability checking.
// This contract formalizes how stock availability should be interpreted.
type StockCheckerContract struct {
	// Available indicates whether the SKU exists and is available for ordering.
	// true = SKU exists and can be ordered
	// false = SKU does not exist or is not available for ordering
	Available bool
	
	// StockQuantity is the actual quantity available in stock.
	// This is the physical quantity on hand.
	// May be 0 even if Available is true (item exists but out of stock).
	StockQuantity uint32
	
	// Error indicates any error that occurred during stock check.
	Error error
}

// InterpretStockResult interprets the StockChecker result according to business rules.
// This method formalizes how we handle different stock scenarios:
//   - Available=false: SKU doesn't exist or is unavailable → Error
//   - Available=true, StockQuantity=0: SKU exists but out of stock → Error
//   - Available=true, StockQuantity < requested: Partial availability → Warning + adjust quantity
//   - Available=true, StockQuantity >= requested: Full availability → Success
func InterpretStockResult(
	available bool,
	stockQuantity uint32,
	requestedQuantity int32,
	goodId interface{},
) StockDecision {
	if !available {
		return StockDecision{
			Type:           StockDecisionTypeSKUUnavailable,
			Reason:         "SKU does not exist or is not available for ordering",
			CanProceed:     false,
			QuantityToUse:  0,
		}
	}

	if stockQuantity == 0 {
		return StockDecision{
			Type:          StockDecisionTypeOutOfStock,
			Reason:        "SKU exists but is currently out of stock",
			CanProceed:    false,
			QuantityToUse: 0,
		}
	}

	if uint32(requestedQuantity) > stockQuantity {
		return StockDecision{
			Type:          StockDecisionTypePartialAvailability,
			Reason:        fmt.Sprintf("Requested %d units, but only %d available", requestedQuantity, stockQuantity),
			CanProceed:    true, // Allow partial fulfillment
			QuantityToUse: int32(stockQuantity),
		}
	}

	return StockDecision{
		Type:          StockDecisionTypeFullAvailability,
		Reason:        "Stock is sufficient for the requested quantity",
		CanProceed:    true,
		QuantityToUse: requestedQuantity,
	}
}

// StockDecisionType represents the type of stock availability decision.
type StockDecisionType string

const (
	StockDecisionTypeSKUUnavailable      StockDecisionType = "SKU_UNAVAILABLE"
	StockDecisionTypeOutOfStock          StockDecisionType = "OUT_OF_STOCK"
	StockDecisionTypePartialAvailability StockDecisionType = "PARTIAL_AVAILABILITY"
	StockDecisionTypeFullAvailability    StockDecisionType = "FULL_AVAILABILITY"
)

// StockDecision represents a decision about how to handle stock availability.
type StockDecision struct {
	// Type indicates the type of stock situation
	Type StockDecisionType
	// Reason provides a human-readable explanation
	Reason string
	// CanProceed indicates whether we can proceed with ordering (even if partial)
	CanProceed bool
	// QuantityToUse is the quantity that should be used for the order
	QuantityToUse int32
}

// IsError returns true if this decision represents an error condition.
func (d StockDecision) IsError() bool {
	return !d.CanProceed
}

// IsWarning returns true if this decision represents a warning condition.
func (d StockDecision) IsWarning() bool {
	return d.CanProceed && d.Type == StockDecisionTypePartialAvailability
}

