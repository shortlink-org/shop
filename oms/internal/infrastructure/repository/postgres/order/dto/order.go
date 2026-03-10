package dto

import (
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common"
	"github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/address"
	"github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/location"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/schema/queries"
)

// OrderRow holds DB rows for one order (header + items + delivery) for conversion to domain.
type OrderRow struct {
	Order    queries.OmsOrder
	Items    []queries.GetOrderItemsRow
	Delivery *queries.GetOrderDeliveryInfoRow
}

// ToDomain converts the row to domain aggregate.
func (r *OrderRow) ToDomain() *order.OrderState {
	domainItems := make(order.Items, 0, len(r.Items))
	for _, i := range r.Items {
		domainItems = append(domainItems, order.NewItem(i.GoodID, i.Quantity, i.Price))
	}

	status := stringToOrderStatus(r.Order.Status)
	deliveryInfo := toDeliveryInfoDomain(r.Delivery)
	deliveryStatus := stringToDeliveryStatus(r.Delivery)
	deliveryRequestedAt := deliveryRequestedAt(r.Delivery)

	return order.NewOrderStateFromPersisted(
		r.Order.ID, r.Order.CustomerID, domainItems,
		status, int(r.Order.Version), deliveryInfo, deliveryStatus, deliveryRequestedAt,
	)
}

// toDeliveryInfoDomain converts database delivery info row to domain DeliveryInfo.
func toDeliveryInfoDomain(row *queries.GetOrderDeliveryInfoRow) *order.DeliveryInfo {
	if row == nil {
		return nil
	}

	// Build pickup address
	pickupLoc, err := location.NewLocation(
		numericToFloat64(row.PickupLatitude),
		numericToFloat64(row.PickupLongitude),
	)
	if err != nil {
		return nil
	}

	pickupAddr, err := address.NewAddressWithLocation(
		row.PickupStreet.String,
		row.PickupCity.String,
		row.PickupPostalCode.String,
		row.PickupCountry.String,
		pickupLoc,
	)
	if err != nil {
		return nil
	}

	// Build delivery address
	deliveryLoc, err := location.NewLocation(
		numericToFloat64(row.DeliveryLatitude),
		numericToFloat64(row.DeliveryLongitude),
	)
	if err != nil {
		return nil
	}

	deliveryAddr, err := address.NewAddressWithLocation(
		row.DeliveryStreet,
		row.DeliveryCity,
		row.DeliveryPostalCode.String,
		row.DeliveryCountry,
		deliveryLoc,
	)
	if err != nil {
		return nil
	}

	// Build delivery period
	period := order.NewDeliveryPeriod(
		row.PeriodStart.Time,
		row.PeriodEnd.Time,
	)

	// Build package info
	pkgInfo := order.NewPackageInfo(numericToFloat64(row.WeightKg))

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

func deliveryRequestedAt(row *queries.GetOrderDeliveryInfoRow) *time.Time {
	if row == nil || !row.RequestedAt.Valid {
		return nil
	}

	requestedAt := row.RequestedAt.Time

	return &requestedAt
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
	case "CANCELED", "CANCELLED", "ORDER_STATUS_CANCELED", "ORDER_STATUS_CANCELLED": //nolint:misspell // accept both spellings
		return order.OrderStatus_ORDER_STATUS_CANCELED
	default:
		return order.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func stringToDeliveryStatus(row *queries.GetOrderDeliveryInfoRow) commonv1.DeliveryStatus {
	if row == nil {
		return commonv1.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED
	}

	switch row.DeliveryStatus {
	case "ACCEPTED", "DELIVERY_STATUS_ACCEPTED":
		return commonv1.DeliveryStatus_DELIVERY_STATUS_ACCEPTED
	case "ASSIGNED", "DELIVERY_STATUS_ASSIGNED":
		return commonv1.DeliveryStatus_DELIVERY_STATUS_ASSIGNED
	case "IN_TRANSIT", "DELIVERY_STATUS_IN_TRANSIT":
		return commonv1.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT
	case "DELIVERED", "DELIVERY_STATUS_DELIVERED":
		return commonv1.DeliveryStatus_DELIVERY_STATUS_DELIVERED
	case "NOT_DELIVERED", "DELIVERY_STATUS_NOT_DELIVERED":
		return commonv1.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED
	default:
		return commonv1.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED
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
		for range int(abs(n.Exp)) {
			exp.Mul(exp, big.NewFloat(10)) //nolint:mnd // decimal base
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
