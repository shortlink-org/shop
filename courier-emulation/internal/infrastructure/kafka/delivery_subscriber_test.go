package kafka

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/require"
)

type mockOrderAssignmentHandler struct {
	events chan OrderAssignedEvent
	err    error
}

func (m *mockOrderAssignmentHandler) HandleOrderAssigned(_ context.Context, event OrderAssignedEvent) error {
	m.events <- event
	return m.err
}

func TestDeliverySubscriber_ProcessMessages_HandlesJSONAssignedEvent(t *testing.T) {
	t.Parallel()

	handler := &mockOrderAssignmentHandler{events: make(chan OrderAssignedEvent, 1)}
	subscriber := &DeliverySubscriber{
		handler: handler,
		logger:  watermill.NewStdLogger(false, false),
		stopCh:  make(chan struct{}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messages := make(chan *message.Message, 1)
	go subscriber.processMessages(ctx, messages)

	assignedAt := time.Date(2026, time.March, 11, 10, 0, 0, 0, time.UTC)
	occurredAt := assignedAt.Add(30 * time.Second)
	payload, err := json.Marshal(OrderAssignedEvent{
		PackageID:  "pkg-1",
		CourierID:  "courier-1",
		Status:     3,
		AssignedAt: assignedAt,
		PickupAddress: Address{
			Latitude:  52.52,
			Longitude: 13.405,
		},
		DeliveryAddress: Address{
			Latitude:  52.53,
			Longitude: 13.415,
		},
		OccurredAt: occurredAt,
	})
	require.NoError(t, err)

	msg := message.NewMessage(watermill.NewUUID(), payload)
	messages <- msg

	select {
	case event := <-handler.events:
		require.Equal(t, "pkg-1", event.PackageID)
		require.Equal(t, "courier-1", event.CourierID)
		require.Equal(t, assignedAt, event.AssignedAt)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for assigned event")
	}

	select {
	case <-msg.Acked():
	case <-time.After(time.Second):
		t.Fatal("expected message to be acked")
	}
}
