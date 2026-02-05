package v1

import "github.com/google/uuid"

// StockAvailabilityInput is the result of a stock check for one good.
// The use case obtains this via a port (e.g. StockChecker) and passes it to the domain.
// The domain layer does not perform I/O; it only interprets this data.
type StockAvailabilityInput struct {
	GoodID        uuid.UUID
	Available     bool   // SKU exists and can be ordered
	StockQuantity uint32 // Physical quantity on hand
	CheckError    error  // Error that occurred during the stock check, if any
}
