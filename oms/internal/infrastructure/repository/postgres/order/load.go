package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/dto"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/schema/crud"
	"github.com/shortlink-org/shop/oms/pkg/uow"
)

// Load retrieves an order by ID.
// Requires transaction in context (use UnitOfWork.Begin()).
func (s *Store) Load(ctx context.Context, orderID uuid.UUID) (*order.OrderState, error) {
	pgxTx := uow.FromContext(ctx)
	if pgxTx == nil {
		return nil, ErrTransactionRequired
	}

	qtx := s.query.WithTx(pgxTx)

	// Get order header
	row, err := qtx.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}

	// Get order items
	items, err := qtx.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Get delivery info (optional)
	var deliveryInfoRow *crud.OmsOrderDeliveryInfo
	deliveryRow, err := qtx.GetOrderDeliveryInfo(ctx, orderID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		// No delivery info - that's OK (self-pickup)
	} else {
		deliveryInfoRow = &deliveryRow
	}

	return dto.ToDomain(row, items, deliveryInfoRow), nil
}

// ListByCustomer retrieves all orders for a customer.
// Requires transaction in context (use UnitOfWork.Begin()).
func (s *Store) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*order.OrderState, error) {
	pgxTx := uow.FromContext(ctx)
	if pgxTx == nil {
		return nil, ErrTransactionRequired
	}

	qtx := s.query.WithTx(pgxTx)

	rows, err := qtx.ListOrdersByCustomer(ctx, customerID)
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

		// Get delivery info (optional)
		var deliveryInfoRow *crud.OmsOrderDeliveryInfo
		deliveryRow, err := qtx.GetOrderDeliveryInfo(ctx, row.ID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}
		} else {
			deliveryInfoRow = &deliveryRow
		}

		orders = append(orders, dto.ToDomainFromList(row, items, deliveryInfoRow))
	}

	return orders, nil
}
