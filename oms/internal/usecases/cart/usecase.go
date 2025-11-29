/*
OMS UC. Application layer
*/
package cart

import (
	"github.com/authzed/authzed-go/v1"
	logger "github.com/shortlink-org/go-sdk/logger"
	"go.temporal.io/sdk/client"

	"github.com/shortlink-org/shop/oms/internal/infrastructure/index"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/websocket"
)

type UC struct {
	// Common
	log logger.Logger

	// Security
	permission *authzed.Client

	// Temporal
	temporalClient client.Client

	// Index for tracking goods in carts
	goodsIndex *index.CartGoodsIndex

	// Websocket notifier for sending notifications to UI
	notifier *websocket.Notifier
}

// New creates a new cart usecase
func New(log logger.Logger, permissionClient *authzed.Client, temporalClient client.Client) (*UC, error) {
	service := &UC{
		log: log,

		// Security
		permission: permissionClient,

		// Temporal
		temporalClient: temporalClient,

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
