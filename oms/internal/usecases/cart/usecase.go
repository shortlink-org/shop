/*
OMS UC. Application layer

This package implements the Cart use case following the pattern:
  Load -> domain method(s) -> Save

Repository is a storage adapter (infrastructure layer), NOT a use-case.
Business operations belong to domain aggregate methods.
*/
package cart

import (
	"github.com/authzed/authzed-go/v1"
	logger "github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/index"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/websocket"
)

type UC struct {
	// Common
	log logger.Logger

	// Security
	permission *authzed.Client

	// Repository for cart persistence
	cartRepo ports.CartRepository

	// Index for tracking goods in carts
	goodsIndex *index.CartGoodsIndex

	// Websocket notifier for sending notifications to UI
	notifier *websocket.Notifier
}

// New creates a new cart usecase
func New(log logger.Logger, permissionClient *authzed.Client, cartRepo ports.CartRepository) (*UC, error) {
	service := &UC{
		log: log,

		// Security
		permission: permissionClient,

		// Repository
		cartRepo: cartRepo,

		// Index for tracking goods in carts
		goodsIndex: index.NewCartGoodsIndex(),

		// Websocket notifier (can be nil if not initialized)
		notifier: nil,
	}

	return service, nil
}

// SetNotifier sets the websocket notifier
func (uc *UC) SetNotifier(notifier *websocket.Notifier) {
	uc.notifier = notifier
}
