/*
Order UC. Infrastructure layer. RPC Endpoint
*/

package v1

import (
	"github.com/shortlink-org/go-sdk/grpc"
	logger "github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/shop/oms/internal/usecases/order"
)

type OrderRPC struct {
	OrderServiceServer

	// Common
	log logger.Logger

	// Services
	orderService *order.UC
}

func New(runRPCServer *grpc.Server, log logger.Logger, orderService *order.UC) (*OrderRPC, error) {
	server := &OrderRPC{
		// Common
		log: log,

		// Services
		orderService: orderService,
	}

	// Register services
	if runRPCServer != nil {
		RegisterOrderServiceServer(runRPCServer.Server, server)
	}

	return server, nil
}
