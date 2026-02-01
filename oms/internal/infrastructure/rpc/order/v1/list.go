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
	query := list.NewQuery(customerID, statusFilter, page, pageSize)
	result, err := o.listHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	// Convert domain orders to proto
	protoOrders := make([]*v1.OrderState, len(result.Orders))
	for i, o := range result.Orders {
		protoOrders[i] = dto.DomainToOrderState(o)
	}

	return &v1.ListResponse{
		Orders:     protoOrders,
		TotalCount: int32(result.TotalCount),
		Pagination: &v1.PaginationResponse{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalPages:  result.TotalPages,
		},
	}, nil
}

// Convert domain OrderStatus to proto OrderStatus
func domainStatusToProto(s order.OrderStatus) commonv1.OrderStatus {
	return commonv1.OrderStatus(s)
}
