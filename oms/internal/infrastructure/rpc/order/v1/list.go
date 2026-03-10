package v1

import (
	"context"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/dto"
	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/model/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/rpcmeta"
	"github.com/shortlink-org/shop/oms/internal/usecases/order/query/list"
)

func (o *OrderRPC) List(ctx context.Context, in *v1.ListRequest) (*v1.ListResponse, error) {
	id, err := rpcmeta.CustomerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	customerID := &id

	// Convert status filter
	var statusFilter []order.OrderStatus
	for _, s := range in.GetStatusFilter() {
		statusFilter = append(statusFilter, s)
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
	start := min(max((page-1)*pageSize, 0), totalCount)

	end := min(start+pageSize, totalCount)

	pageOrders := orders[start:end]

	// Convert domain orders to proto
	protoOrders := make([]*v1.OrderState, len(pageOrders))
	for i, o := range pageOrders {
		protoOrders[i] = dto.DomainToOrderState(o)
	}

	totalPages := max((totalCount+pageSize-1)/pageSize, 1)

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
