package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/dto"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/schema/crud"
	"github.com/shortlink-org/shop/oms/pkg/uow"
)

// List retrieves orders with filtering and pagination.
// Requires transaction in context (use UnitOfWork.Begin()).
func (s *Store) List(ctx context.Context, filter ports.ListFilter, page, pageSize int32) (*ports.ListResult, error) {
	pgxTx := uow.FromContext(ctx)
	if pgxTx == nil {
		return nil, ErrTransactionRequired
	}

	qtx := s.query.WithTx(pgxTx)

	// Calculate offset
	offset := (page - 1) * pageSize

	var rows []crud.OmsOrder
	var totalCount int64
	var err error

	// Determine which query to use based on filters
	hasCustomer := filter.CustomerID != nil
	hasStatus := len(filter.StatusFilter) > 0

	if hasCustomer && hasStatus {
		// Both filters
		statusInts := statusesToInts(filter.StatusFilter)
		rows, err = qtx.ListOrdersWithFilters(ctx, crud.ListOrdersWithFiltersParams{
			CustomerID: *filter.CustomerID,
			Column2:    statusInts,
			Limit:      pageSize,
			Offset:     offset,
		})
		if err != nil {
			return nil, err
		}
		totalCount, err = qtx.CountOrdersWithFilters(ctx, crud.CountOrdersWithFiltersParams{
			CustomerID: *filter.CustomerID,
			Column2:    statusInts,
		})
	} else if hasCustomer {
		// Customer filter only
		rows, err = qtx.ListOrdersWithCustomerFilter(ctx, crud.ListOrdersWithCustomerFilterParams{
			CustomerID: *filter.CustomerID,
			Limit:      pageSize,
			Offset:     offset,
		})
		if err != nil {
			return nil, err
		}
		totalCount, err = qtx.CountOrdersByCustomer(ctx, *filter.CustomerID)
	} else if hasStatus {
		// Status filter only
		statusInts := statusesToInts(filter.StatusFilter)
		rows, err = qtx.ListOrdersWithStatusFilter(ctx, crud.ListOrdersWithStatusFilterParams{
			Column1: statusInts,
			Limit:   pageSize,
			Offset:  offset,
		})
		if err != nil {
			return nil, err
		}
		totalCount, err = qtx.CountOrdersByStatus(ctx, statusInts)
	} else {
		// No filters
		rows, err = qtx.ListOrders(ctx, crud.ListOrdersParams{
			Limit:  pageSize,
			Offset: offset,
		})
		if err != nil {
			return nil, err
		}
		totalCount, err = qtx.CountOrders(ctx)
	}

	if err != nil {
		return nil, err
	}

	// Convert rows to domain orders
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

	// Calculate total pages
	totalPages := int32(totalCount / int64(pageSize))
	if totalCount%int64(pageSize) > 0 {
		totalPages++
	}

	return &ports.ListResult{
		Orders:     orders,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}, nil
}

// statusesToInts converts OrderStatus slice to int32 slice for SQL queries.
func statusesToInts(statuses []order.OrderStatus) []int32 {
	result := make([]int32, len(statuses))
	for i, s := range statuses {
		result[i] = int32(s)
	}
	return result
}
