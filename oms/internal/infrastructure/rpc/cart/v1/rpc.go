/*
Cart UC. Infrastructure layer. RPC Endpoint
*/

package v1

import (
	"github.com/shortlink-org/go-sdk/grpc"
	logger "github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/oms/internal/usecases/cart/command/add_items"
	"github.com/shortlink-org/shop/oms/internal/usecases/cart/command/remove_items"
	"github.com/shortlink-org/shop/oms/internal/usecases/cart/command/reset"
	"github.com/shortlink-org/shop/oms/internal/usecases/cart/query/get"
)

type CartRPC struct {
	CartServiceServer

	// Common
	log logger.Logger

	// Command Handlers (concrete types for wire compatibility)
	addItemsHandler    *add_items.Handler
	removeItemsHandler *remove_items.Handler
	resetHandler       *reset.Handler

	// Query Handlers
	getHandler *get.Handler
}

func New(
	runRPCServer *grpc.Server,
	log logger.Logger,
	addItemsHandler *add_items.Handler,
	removeItemsHandler *remove_items.Handler,
	resetHandler *reset.Handler,
	getHandler *get.Handler,
) (*CartRPC, error) {
	server := &CartRPC{
		// Common
		log: log,

		// Command Handlers
		addItemsHandler:    addItemsHandler,
		removeItemsHandler: removeItemsHandler,
		resetHandler:       resetHandler,

		// Query Handlers
		getHandler: getHandler,
	}

	// Register services
	if runRPCServer != nil {
		RegisterCartServiceServer(runRPCServer.Server, server)
	}

	return server, nil
}
