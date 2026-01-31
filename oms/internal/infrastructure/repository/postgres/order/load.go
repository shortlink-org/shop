package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/dto"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/tx"
)

// Load retrieves an order by ID.
// Requires transaction in context (use UnitOfWork.Begin()).
func (s *Store) Load(ctx context.Context, orderID uuid.UUID) (*order.OrderState, error) {
	pgxTx := tx.FromContext(ctx)
	if pgxTx == nil {
		return nil, ErrTransactionRequired
	}

	qtx := s.query.WithTx(pgxTx)

	// Get order header
	row, err := qtx.GetOrder(ctx, uuidToPgtype(orderID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}

	// Get order items
	items, err := qtx.GetOrderItems(ctx, uuidToPgtype(orderID))
	if err != nil {
		return nil, err
	}

	return dto.ToDomain(row, items), nil
}

// ListByCustomer retrieves all orders for a customer.
// Requires transaction in context (use UnitOfWork.Begin()).
func (s *Store) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*order.OrderState, error) {
	pgxTx := tx.FromContext(ctx)
	if pgxTx == nil {
		return nil, ErrTransactionRequired
	}

	qtx := s.query.WithTx(pgxTx)

	rows, err := qtx.ListOrdersByCustomer(ctx, uuidToPgtype(customerID))
	if err != nil {
		return nil, err
	}

	orders := make([]*order.OrderState, 0, len(rows))
	for _, row := range rows {
		// Get items for each order
		items, err := qtx.GetOrderItems(ctx, row.ID)
		if err != nil {
			return nil, err
		}

		orders = append(orders, dto.ToDomainFromList(row, items))
	}

	return orders, nil
}

// pgtypeUUIDToUUID converts pgtype.UUID to uuid.UUID
func pgtypeUUIDToUUID(p pgtype.UUID) uuid.UUID {
	if !p.Valid {
		return uuid.Nil
	}
	return uuid.UUID(p.Bytes)
}
