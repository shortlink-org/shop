package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/schema/queries"
	"github.com/shortlink-org/shop/oms/pkg/uow"
)

// ErrTransactionRequired is returned when repository is called without UoW transaction.
var ErrTransactionRequired = errors.New("transaction required: use UnitOfWork.Begin()")

// Save persists the order state with optimistic concurrency control.
// Requires transaction in context (use UnitOfWork.Begin()).
func (s *Store) Save(ctx context.Context, state *order.OrderState) error {
	pgxTx := uow.FromContext(ctx)
	if pgxTx == nil {
		return ErrTransactionRequired
	}

	qtx := s.query.WithTx(pgxTx)

	orderID := state.GetOrderID()
	customerID := state.GetCustomerId()
	status := state.GetStatus().String()
	newVersion := int32(state.GetVersion() + 1)
	oldVersion := int32(state.GetVersion())

	if oldVersion == 0 {
		// New order - insert
		if err := qtx.InsertOrder(ctx, queries.InsertOrderParams{
			ID:         orderID,
			CustomerID: customerID,
			Status:     status,
		}); err != nil {
			return err
		}
	} else {
		// Update with optimistic lock
		result, err := qtx.UpdateOrder(ctx, queries.UpdateOrderParams{
			ID:        orderID,
			Status:    status,
			Version:   newVersion,
			Version_2: oldVersion,
		})
		if err != nil {
			return err
		}

		if result.RowsAffected() == 0 {
			return ports.ErrVersionConflict
		}
	}

	// Delete existing items and insert new ones
	if err := qtx.DeleteOrderItems(ctx, orderID); err != nil {
		return err
	}

	for _, item := range state.GetItems() {
		if err := qtx.InsertOrderItem(ctx, queries.InsertOrderItemParams{
			OrderID:  orderID,
			GoodID:   item.GetGoodId(),
			Quantity: item.GetQuantity(),
			Price:    item.GetPrice(),
		}); err != nil {
			return err
		}
	}

	// Save delivery info if present
	if err := s.saveDeliveryInfo(ctx, qtx, orderID, state.GetDeliveryInfo(), oldVersion == 0); err != nil {
		return err
	}

	// Invalidate L1 cache after successful save
	s.invalidateCache(orderID.String())

	return nil
}

// saveDeliveryInfo saves or updates delivery info for an order.
func (s *Store) saveDeliveryInfo(ctx context.Context, qtx *queries.Queries, orderID uuid.UUID, deliveryInfo *order.DeliveryInfo, isNew bool) error {
	if deliveryInfo == nil {
		// No delivery info - delete if exists (for updates)
		if !isNew {
			return qtx.DeleteOrderDeliveryInfo(ctx, orderID)
		}
		return nil
	}

	pickupAddr := deliveryInfo.GetPickupAddress()
	deliveryAddr := deliveryInfo.GetDeliveryAddress()
	period := deliveryInfo.GetDeliveryPeriod()
	pkgInfo := deliveryInfo.GetPackageInfo()

	// Convert package ID to pgtype.UUID
	var packageID pgtype.UUID
	if pkgID := deliveryInfo.GetPackageId(); pkgID != nil {
		packageID = pgtype.UUID{Bytes: *pkgID, Valid: true}
	}

	// Recipient contacts (optional)
	var recipientName, recipientPhone, recipientEmail pgtype.Text
	if rc := deliveryInfo.GetRecipientContacts(); rc != nil {
		recipientName = pgtype.Text{String: rc.GetName(), Valid: rc.GetName() != ""}
		recipientPhone = pgtype.Text{String: rc.GetPhone(), Valid: rc.GetPhone() != ""}
		recipientEmail = pgtype.Text{String: rc.GetEmail(), Valid: rc.GetEmail() != ""}
	}

	params := queries.InsertOrderDeliveryInfoParams{
		OrderID:            orderID,
		PickupStreet:       pgtype.Text{String: pickupAddr.Street(), Valid: pickupAddr.Street() != ""},
		PickupCity:         pgtype.Text{String: pickupAddr.City(), Valid: pickupAddr.City() != ""},
		PickupPostalCode:   pgtype.Text{String: pickupAddr.PostalCode(), Valid: pickupAddr.PostalCode() != ""},
		PickupCountry:      pgtype.Text{String: pickupAddr.Country(), Valid: pickupAddr.Country() != ""},
		PickupLatitude:     float64ToNumeric(pickupAddr.Latitude()),
		PickupLongitude:    float64ToNumeric(pickupAddr.Longitude()),
		DeliveryStreet:     deliveryAddr.Street(),
		DeliveryCity:       deliveryAddr.City(),
		DeliveryPostalCode: pgtype.Text{String: deliveryAddr.PostalCode(), Valid: deliveryAddr.PostalCode() != ""},
		DeliveryCountry:    deliveryAddr.Country(),
		DeliveryLatitude:   float64ToNumeric(deliveryAddr.Latitude()),
		DeliveryLongitude:  float64ToNumeric(deliveryAddr.Longitude()),
		PeriodStart:        pgtype.Timestamptz{Time: period.GetStartTime(), Valid: true},
		PeriodEnd:          pgtype.Timestamptz{Time: period.GetEndTime(), Valid: true},
		WeightKg:   float64ToNumeric(pkgInfo.GetWeightKg()),
		Priority:   deliveryInfo.GetPriority().String(),
		PackageID:          packageID,
		RecipientName:      recipientName,
		RecipientPhone:     recipientPhone,
		RecipientEmail:     recipientEmail,
	}

	if isNew {
		return qtx.InsertOrderDeliveryInfo(ctx, params)
	}

	// For updates, delete and re-insert (simpler than upsert)
	if err := qtx.DeleteOrderDeliveryInfo(ctx, orderID); err != nil {
		return err
	}
	return qtx.InsertOrderDeliveryInfo(ctx, params)
}

// invalidateCache removes an order from the L1 cache.
func (s *Store) invalidateCache(orderID string) {
	s.cache.Del(orderID)
}

// float64ToNumeric converts a float64 to pgtype.Numeric.
func float64ToNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(f)
	return n
}
