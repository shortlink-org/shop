package v1

import (
	"context"

	"github.com/google/uuid"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/dto"
	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/model/v1"
	"github.com/shortlink-org/shop/oms/internal/usecases/order/query/list"
)

func (o *OrderRPC) List(ctx context.Context, in *v1.ListRequest) (*v1.ListResponse, error) {
	// Parse optional customer ID
	var customerID *uuid.UUID
	if in.GetCustomerId() != "" {
		id, err := uuid.Parse(in.GetCustomerId())
		if err != nil {
			return nil, err
		}
		customerID = &id
	}

	// Convert status filter
	var statusFilter []order.OrderStatus
	for _, s := range in.GetStatusFilter() {
		statusFilter = append(statusFilter, order.OrderStatus(s))
	}

	// Get pagination
	page := int32(1)
	pageSize := int32(20)
	if in.GetPagination() != nil {
		if in.GetPagination().GetPage() > 0 {
			page = in.GetPagination().GetPage()
		}
		if in.GetPagination().GetPageSize() > 0 {
			pageSize = in.GetPagination().GetPageSize()
		}
	}

	// Create query and execute handler
	query := list.NewQuery(customerID, statusFilter)
	orders, err := o.listHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	totalCount := int32(len(orders))
	// Paginate in RPC layer
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	if start >= totalCount {
		start = totalCount
	}
	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}
	pageOrders := orders[start:end]

	// Convert domain orders to proto
	protoOrders := make([]*v1.OrderState, len(pageOrders))
	for i, o := range pageOrders {
		protoOrders[i] = dto.DomainToOrderState(o)
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	return &v1.ListResponse{
		Orders:     protoOrders,
		TotalCount: totalCount,
		Pagination: &v1.PaginationResponse{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalPages:  totalPages,
		},
	}, nil
}

// Convert domain OrderStatus to proto OrderStatus
func domainStatusToProto(s order.OrderStatus) commonv1.OrderStatus {
	return commonv1.OrderStatus(s)
}
