package on_stock_changed

import (
	"time"

	"github.com/google/uuid"
)

// Event represents a fact: stock level changed in inventory.
// This is NOT a command (intent), but an event (fact that already happened).
type Event struct {
	GoodID      uuid.UUID
	NewQuantity uint32
	OccurredAt  time.Time
}

// EventType returns the event type name for the event publisher.
func (e Event) EventType() string {
	return "StockChanged"
}

// NewEvent creates a new StockChangedEvent.
func NewEvent(goodID uuid.UUID, newQuantity uint32) Event {
	return Event{
		GoodID:      goodID,
		NewQuantity: newQuantity,
		OccurredAt:  time.Now(),
	}
}
