package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/schema/crud"
)

// Save persists the order state with optimistic concurrency control.
func (s *Store) Save(ctx context.Context, state *order.OrderState) error {
	tx, err := s.client.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := s.query.WithTx(tx)

	orderID := uuidToPgtype(state.GetOrderID())
	customerID := uuidToPgtype(state.GetCustomerId())
	status := state.GetStatus().String()
	newVersion := int32(state.GetVersion() + 1)
	oldVersion := int32(state.GetVersion())

	if oldVersion == 0 {
		// New order - insert
		if err := qtx.InsertOrder(ctx, crud.InsertOrderParams{
			ID:         orderID,
			CustomerID: customerID,
			Status:     status,
		}); err != nil {
			return err
		}
	} else {
		// Update with optimistic lock
		result, err := qtx.UpdateOrder(ctx, crud.UpdateOrderParams{
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
		if err := qtx.InsertOrderItem(ctx, crud.InsertOrderItemParams{
			OrderID:  orderID,
			GoodID:   uuidToPgtype(item.GetGoodId()),
			Quantity: item.GetQuantity(),
			Price:    decimalToPgtype(item.GetPrice()),
		}); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// uuidToPgtype converts uuid.UUID to pgtype.UUID
func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}

// decimalToPgtype converts decimal to pgtype.Numeric
func decimalToPgtype(d interface{ String() string }) pgtype.Numeric {
	var num pgtype.Numeric
	_ = num.Scan(d.String())
	return num
}
