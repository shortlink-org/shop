/*
Checkout UseCase. Application layer

This package implements the Checkout use case which atomically:
  - Creates an order from cart items
  - Clears the cart
  - Uses UnitOfWork for transactional consistency

All operations are wrapped in a single database transaction via UoW.
*/
package checkout

import (
	logger "github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
)

type UC struct {
	log       logger.Logger
	uow       ports.UnitOfWork
	cartRepo  ports.CartRepository
	orderRepo ports.OrderRepository
}

// New creates a new checkout usecase
func New(
	log logger.Logger,
	uow ports.UnitOfWork,
	cartRepo ports.CartRepository,
	orderRepo ports.OrderRepository,
) *UC {
	return &UC{
		log:       log,
		uow:       uow,
		cartRepo:  cartRepo,
		orderRepo: orderRepo,
	}
}
