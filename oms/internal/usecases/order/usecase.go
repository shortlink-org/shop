/*
OMS UC. Application layer

This package implements the Order use case following the pattern:
  Load -> domain method(s) -> Save

For complex order workflows (sagas), Temporal is used for orchestration.
Repository handles persistence, Temporal handles workflow coordination.
*/
package order

import (
	"github.com/authzed/authzed-go/v1"
	logger "github.com/shortlink-org/go-sdk/logger"
	"go.temporal.io/sdk/client"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
)

type UC struct {
	// Common
	log logger.Logger

	// Security
	permission *authzed.Client

	// Repository for order persistence
	orderRepo ports.OrderRepository

	// Temporal for workflow orchestration (sagas, long-running processes)
	temporalClient client.Client
}

// New creates a new order usecase
func New(log logger.Logger, permissionClient *authzed.Client, orderRepo ports.OrderRepository, temporalClient client.Client) (*UC, error) {
	service := &UC{
		log: log,

		// Security
		permission: permissionClient,

		// Repository
		orderRepo: orderRepo,

		// Temporal for workflow orchestration
		temporalClient: temporalClient,
	}

	return service, nil
}
