package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
	cart "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/cart/schema/crud"
)

// Save persists the cart state with optimistic concurrency control.
func (s *Store) Save(ctx context.Context, state *cart.State) error {
	tx, err := s.client.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := s.query.WithTx(tx)

	customerID := uuidToPgtype(state.GetCustomerId())
	newVersion := int32(state.GetVersion() + 1)
	oldVersion := int32(state.GetVersion())

	// Try to update with optimistic lock
	if oldVersion > 0 {
		result, err := qtx.UpsertCart(ctx, crud.UpsertCartParams{
			CustomerID: customerID,
			Version:    newVersion,
			Version_2:  oldVersion,
		})
		if err != nil {
			return err
		}

		rowsAffected := result.RowsAffected()
		if rowsAffected == 0 {
			return ports.ErrVersionConflict
		}
	} else {
		// New cart - insert
		if err := qtx.InsertCart(ctx, customerID); err != nil {
			return err
		}
	}

	// Delete existing items and insert new ones
	if err := qtx.DeleteCartItems(ctx, customerID); err != nil {
		return err
	}

	for _, item := range state.GetItems() {
		if err := qtx.InsertCartItem(ctx, crud.InsertCartItemParams{
			CartID:   customerID,
			GoodID:   uuidToPgtype(item.GetGoodId()),
			Quantity: item.GetQuantity(),
			Price:    decimalToPgtype(item.GetPrice()),
			Discount: decimalToPgtype(item.GetDiscount()),
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

// decimalToPgtype converts shopspring decimal to pgtype.Numeric
func decimalToPgtype(d interface{ String() string }) pgtype.Numeric {
	var num pgtype.Numeric
	_ = num.Scan(d.String())
	return num
}
