package events

// Event is the canonical domain event contract shared by aggregates.
type Event interface {
	EventType() string
}
