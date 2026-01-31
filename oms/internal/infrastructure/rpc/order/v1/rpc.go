/*
Order UC. Infrastructure layer. RPC Endpoint
*/

package v1

import (
	"github.com/shortlink-org/go-sdk/grpc"
	logger "github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/oms/internal/usecases/checkout/command/create_order_from_cart"
	"github.com/shortlink-org/shop/oms/internal/usecases/order/command/cancel"
	"github.com/shortlink-org/shop/oms/internal/usecases/order/command/create"
	"github.com/shortlink-org/shop/oms/internal/usecases/order/command/update_delivery_info"
	"github.com/shortlink-org/shop/oms/internal/usecases/order/query/get"
)

type OrderRPC struct {
	OrderServiceServer

	// Common
	log logger.Logger

	// Command Handlers (concrete types for wire compatibility)
	createHandler             *create.Handler
	cancelHandler             *cancel.Handler
	updateDeliveryInfoHandler *update_delivery_info.Handler
	checkoutHandler           *create_order_from_cart.Handler

	// Query Handlers
	getHandler *get.Handler
}

func New(
	runRPCServer *grpc.Server,
	log logger.Logger,
	createHandler *create.Handler,
	cancelHandler *cancel.Handler,
	updateDeliveryInfoHandler *update_delivery_info.Handler,
	checkoutHandler *create_order_from_cart.Handler,
	getHandler *get.Handler,
) (*OrderRPC, error) {
	server := &OrderRPC{
		// Common
		log: log,

		// Command Handlers
		createHandler:             createHandler,
		cancelHandler:             cancelHandler,
		updateDeliveryInfoHandler: updateDeliveryInfoHandler,
		checkoutHandler:           checkoutHandler,

		// Query Handlers
		getHandler: getHandler,
	}

	// Register services
	if runRPCServer != nil {
		RegisterOrderServiceServer(runRPCServer.Server, server)
	}

	return server, nil
}
