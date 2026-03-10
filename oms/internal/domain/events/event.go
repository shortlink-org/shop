package events

// Event is the canonical domain event contract shared by aggregates.
// Implementations provide EventType() for routing (e.g. in-memory or Kafka).
//
//nolint:iface // contract type used by other packages
type Event interface {
	EventType() string
}
