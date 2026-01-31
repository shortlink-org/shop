package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
	cart "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/cart/schema/crud"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/tx"
)

// ErrTransactionRequired is returned when repository is called without UoW transaction.
var ErrTransactionRequired = errors.New("transaction required: use UnitOfWork.Begin()")

// Save persists the cart state with optimistic concurrency control.
// Requires transaction in context (use UnitOfWork.Begin()).
func (s *Store) Save(ctx context.Context, state *cart.State) error {
	pgxTx := tx.FromContext(ctx)
	if pgxTx == nil {
		return ErrTransactionRequired
	}

	qtx := s.query.WithTx(pgxTx)

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

	return nil
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
