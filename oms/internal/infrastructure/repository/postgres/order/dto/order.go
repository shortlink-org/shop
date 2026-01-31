package dto

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/schema/crud"
)

// ToDomain converts database models to domain aggregate.
func ToDomain(row crud.OmsOrder, items []crud.GetOrderItemsRow) *order.OrderState {
	domainItems := make(order.Items, 0, len(items))

	for _, i := range items {
		goodID := pgtypeUUIDToUUID(i.GoodID)
		price := pgtypeNumericToDecimal(i.Price)

		item := order.NewItem(goodID, i.Quantity, price)
		domainItems = append(domainItems, item)
	}

	id := pgtypeUUIDToUUID(row.ID)
	customerID := pgtypeUUIDToUUID(row.CustomerID)
	status := stringToOrderStatus(row.Status)

	return order.Reconstitute(id, customerID, domainItems, status, int(row.Version))
}

// ToDomainFromList converts database models from list query to domain aggregate.
func ToDomainFromList(row crud.OmsOrder, items []crud.GetOrderItemsRow) *order.OrderState {
	return ToDomain(row, items)
}

// stringToOrderStatus converts status string to OrderStatus enum.
func stringToOrderStatus(s string) order.OrderStatus {
	switch s {
	case "PENDING", "ORDER_STATUS_PENDING":
		return order.OrderStatus_ORDER_STATUS_PENDING
	case "PROCESSING", "ORDER_STATUS_PROCESSING":
		return order.OrderStatus_ORDER_STATUS_PROCESSING
	case "COMPLETED", "ORDER_STATUS_COMPLETED":
		return order.OrderStatus_ORDER_STATUS_COMPLETED
	case "CANCELLED", "ORDER_STATUS_CANCELLED":
		return order.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return order.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

// pgtypeUUIDToUUID converts pgtype.UUID to uuid.UUID
func pgtypeUUIDToUUID(p pgtype.UUID) uuid.UUID {
	if !p.Valid {
		return uuid.Nil
	}
	return uuid.UUID(p.Bytes)
}

// pgtypeNumericToDecimal converts pgtype.Numeric to decimal.Decimal
func pgtypeNumericToDecimal(p pgtype.Numeric) decimal.Decimal {
	if !p.Valid {
		return decimal.Zero
	}

	// Convert pgtype.Numeric to float64
	f, err := p.Float64Value()
	if err == nil && f.Valid {
		return decimal.NewFromFloat(f.Float64)
	}

	// Fallback: try Int64
	i, err := p.Int64Value()
	if err == nil && i.Valid {
		return decimal.NewFromInt(i.Int64)
	}

	return decimal.Zero
}
