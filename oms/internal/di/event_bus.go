package oms_di

import (
	"github.com/ThreeDotsLabs/watermill"

	"github.com/shortlink-org/go-sdk/cqrs/bus"
	cqrsmessage "github.com/shortlink-org/go-sdk/cqrs/message"
	logger "github.com/shortlink-org/go-sdk/logger"
)

const outboxForwarderTopic = "oms_outbox"

// newEventBus creates EventBus with tx-aware outbox (go-sdk/uow).
// Publisher is nil: Publish must be called inside UoW transaction or it returns an error.
func newEventBus(log logger.Logger) *bus.EventBus {
	namer := cqrsmessage.NewShortlinkNamer("oms")
	marshaler := cqrsmessage.NewJSONMarshaler(namer)
	wmLogger := watermill.NewStdLogger(false, false)
	return bus.NewEventBus(
		nil,
		marshaler,
		namer,
		bus.WithTxAwareOutbox(outboxForwarderTopic, wmLogger),
	)
}
