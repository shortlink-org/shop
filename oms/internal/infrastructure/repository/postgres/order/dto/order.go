package dto

import (
	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/schema/crud"
)

// ToDomain converts database models to domain aggregate.
func ToDomain(row crud.OmsOrder, items []crud.GetOrderItemsRow) *order.OrderState {
	domainItems := make(order.Items, 0, len(items))

	for _, i := range items {
		item := order.NewItem(i.GoodID, i.Quantity, i.Price)
		domainItems = append(domainItems, item)
	}

	status := stringToOrderStatus(row.Status)

	return order.Reconstitute(row.ID, row.CustomerID, domainItems, status, int(row.Version))
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
