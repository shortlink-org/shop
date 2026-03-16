package ports

import "context"

// DeliveryInboxRepository deduplicates inbound delivery events from Kafka.
//
//nolint:iface // port interface used by usecases and DI
type DeliveryInboxRepository interface {
	TryRecord(ctx context.Context, consumerName, messageID, topic string) (bool, error)
}
