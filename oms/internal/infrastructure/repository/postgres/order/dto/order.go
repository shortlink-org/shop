package dto

import (
	"math/big"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/address"
	"github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/location"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/schema/queries"
)

// OrderRow holds DB rows for one order (header + items + delivery) for conversion to domain.
type OrderRow struct {
	Order    queries.OmsOrder
	Items    []queries.GetOrderItemsRow
	Delivery *queries.OmsOrderDeliveryInfo
}

// ToDomain converts the row to domain aggregate.
func (r *OrderRow) ToDomain() *order.OrderState {
	domainItems := make(order.Items, 0, len(r.Items))
	for _, i := range r.Items {
		domainItems = append(domainItems, order.NewItem(i.GoodID, i.Quantity, i.Price))
	}
	status := stringToOrderStatus(r.Order.Status)
	deliveryInfo := toDeliveryInfoDomain(r.Delivery)
	return order.NewOrderStateFromPersisted(
		r.Order.ID, r.Order.CustomerID, domainItems,
		status, int(r.Order.Version), deliveryInfo,
	)
}

// toDeliveryInfoDomain converts database delivery info row to domain DeliveryInfo.
func toDeliveryInfoDomain(row *queries.OmsOrderDeliveryInfo) *order.DeliveryInfo {
	if row == nil {
		return nil
	}

	// Build pickup address
	pickupLoc, _ := location.NewLocation(
		numericToFloat64(row.PickupLatitude),
		numericToFloat64(row.PickupLongitude),
	)
	pickupAddr, _ := address.NewAddressWithLocation(
		row.PickupStreet.String,
		row.PickupCity.String,
		row.PickupPostalCode.String,
		row.PickupCountry.String,
		pickupLoc,
	)

	// Build delivery address
	deliveryLoc, _ := location.NewLocation(
		numericToFloat64(row.DeliveryLatitude),
		numericToFloat64(row.DeliveryLongitude),
	)
	deliveryAddr, _ := address.NewAddressWithLocation(
		row.DeliveryStreet,
		row.DeliveryCity,
		row.DeliveryPostalCode.String,
		row.DeliveryCountry,
		deliveryLoc,
	)

	// Build delivery period
	period := order.NewDeliveryPeriod(
		row.PeriodStart.Time,
		row.PeriodEnd.Time,
	)

	// Build package info
	pkgInfo := order.NewPackageInfo(
		numericToFloat64(row.WeightKg),
		row.Dimensions.String,
	)

	// Build priority
	priority := order.DeliveryPriorityFromString(row.Priority)

	// Build optional recipient contacts
	var recipientContacts *order.RecipientContacts
	if row.RecipientName.Valid || row.RecipientPhone.Valid || row.RecipientEmail.Valid {
		rc := order.NewRecipientContacts(
			row.RecipientName.String,
			row.RecipientPhone.String,
			row.RecipientEmail.String,
		)
		recipientContacts = &rc
	}

	// Create delivery info
	deliveryInfo := order.NewDeliveryInfo(
		pickupAddr,
		deliveryAddr,
		period,
		pkgInfo,
		priority,
		recipientContacts,
	)

	// Set package ID if present
	if row.PackageID.Valid {
		pkgID := uuid.UUID(row.PackageID.Bytes)
		deliveryInfo.SetPackageId(pkgID)
	}

	return &deliveryInfo
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

// numericToFloat64 converts pgtype.Numeric to float64.
func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}

	// Convert to big.Float for proper handling
	bigFloat := new(big.Float)
	bigFloat.SetInt(n.Int)

	// Apply exponent
	if n.Exp != 0 {
		exp := new(big.Float).SetFloat64(1)
		for i := int32(0); i < abs(n.Exp); i++ {
			exp.Mul(exp, big.NewFloat(10))
		}
		if n.Exp < 0 {
			bigFloat.Quo(bigFloat, exp)
		} else {
			bigFloat.Mul(bigFloat, exp)
		}
	}

	result, _ := bigFloat.Float64()
	return result
}

// abs returns absolute value of int32.
func abs(n int32) int32 {
	if n < 0 {
		return -n
	}
	return n
}
