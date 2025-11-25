/*
Cart UC. Infrastructure layer. RPC Endpoint
*/

package v1

import (
	logger "github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/shop/oms/internal/usecases/cart"
	"github.com/shortlink-org/shortlink/pkg/rpc"
)

type CartRPC struct {
	CartServiceServer

	// Common
	log logger.Logger

	// Services
	cartService *cart.UC
}

func New(runRPCServer *rpc.Server, log logger.Logger, cartService *cart.UC) (*CartRPC, error) {
	server := &CartRPC{
		// Common
		log: log,

		// Services
		cartService: cartService,
	}

	// Register services
	if runRPCServer != nil {
		RegisterCartServiceServer(runRPCServer.Server, server)
	}

	return server, nil
}
