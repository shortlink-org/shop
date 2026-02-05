package kafka

// Topic name format: {domain}.{entity}.{event}.v1
const (
	topicDomain = "delivery"
	topicEntity = "order"
	topicSuffix = ".v1"

	eventNameOrderPickedUp  = "order_picked_up"
	eventNameOrderDelivered = "order_delivered"

	topicPrefix = topicDomain + "." + topicEntity + "."

	// TopicPickUpOrder is the Kafka topic for order picked up events.
	TopicPickUpOrder = topicPrefix + eventNameOrderPickedUp + topicSuffix
	// TopicDeliverOrder is the Kafka topic for order delivered events.
	TopicDeliverOrder = topicPrefix + eventNameOrderDelivered + topicSuffix
)

// Metadata keys for Kafka messages.
const (
	metadataKeyPartitionKey = "partition_key"
)
