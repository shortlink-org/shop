package postgres

import (
	"context"
	"errors"
	"math"

	"github.com/jackc/pgx/v5"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/dto"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/schema/queries"
	"github.com/shortlink-org/shop/oms/pkg/uow"
)

// List retrieves orders with optional filter. No pagination.
// Requires transaction in context (use UnitOfWork.Begin()).
func (s *Store) List(ctx context.Context, filter ports.ListFilter) ([]*order.OrderState, error) {
	pgxTx := uow.FromContext(ctx)
	if pgxTx == nil {
		return nil, ErrTransactionRequired
	}

	qtx := s.query.WithTx(pgxTx)

	var (
		rows []queries.OmsOrder
		err  error
	)

	hasCustomer := filter.CustomerID != nil
	hasStatus := len(filter.StatusFilter) > 0

	if hasCustomer && hasStatus {
		statusInts := statusesToInts(filter.StatusFilter)
		rows, err = qtx.ListOrdersWithFilters(ctx, queries.ListOrdersWithFiltersParams{
			CustomerID: *filter.CustomerID,
			Column2:    statusInts,
			Limit:      math.MaxInt32,
			Offset:     0,
		})
	} else if hasCustomer {
		rows, err = qtx.ListOrdersWithCustomerFilter(ctx, queries.ListOrdersWithCustomerFilterParams{
			CustomerID: *filter.CustomerID,
			Limit:      math.MaxInt32,
			Offset:     0,
		})
	} else if hasStatus {
		statusInts := statusesToInts(filter.StatusFilter)
		rows, err = qtx.ListOrdersWithStatusFilter(ctx, queries.ListOrdersWithStatusFilterParams{
			Column1: statusInts,
			Limit:   math.MaxInt32,
			Offset:  0,
		})
	} else {
		rows, err = qtx.ListOrders(ctx, queries.ListOrdersParams{
			Limit:  math.MaxInt32,
			Offset: 0,
		})
	}

	if err != nil {
		return nil, err
	}

	orders := make([]*order.OrderState, 0, len(rows))
	for _, row := range rows {
		items, err := qtx.GetOrderItems(ctx, row.ID)
		if err != nil {
			return nil, err
		}

		var deliveryInfoRow *queries.OmsOrderDeliveryInfo

		deliveryRow, err := qtx.GetOrderDeliveryInfo(ctx, row.ID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}
		} else {
			deliveryInfoRow = &deliveryRow
		}

		orders = append(orders, (&dto.OrderRow{Order: row, Items: items, Delivery: deliveryInfoRow}).ToDomain())
	}

	return orders, nil
}

// statusesToInts converts OrderStatus slice to int32 slice for SQL queries.
func statusesToInts(statuses []order.OrderStatus) []int32 {
	result := make([]int32, len(statuses))
	for i, s := range statuses {
		result[i] = int32(s)
	}

	return result
}
