package kafka

import (
	"context"
	"testing"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
)

// mockPublisher is a mock implementation of message.Publisher for testing.
type mockPublisher struct {
	messages map[string][]*message.Message
	closed   bool
}

func newMockPublisher() *mockPublisher {
	return &mockPublisher{
		messages: make(map[string][]*message.Message),
	}
}

func (m *mockPublisher) Publish(topic string, messages ...*message.Message) error {
	m.messages[topic] = append(m.messages[topic], messages...)
	return nil
}

func (m *mockPublisher) Close() error {
	m.closed = true
	return nil
}

func TestLocationPublisher_PublishLocation(t *testing.T) {
	mock := newMockPublisher()
	publisher := &LocationPublisher{
		publisher: mock,
		logger:    watermill.NewStdLogger(false, false),
	}

	location := vo.MustNewLocation(52.5200, 13.4050)
	event := vo.NewCourierLocationEvent("courier-1", location, vo.CourierStatusMoving).
		WithSpeed(25.5).
		WithRouteID("route-001")

	err := publisher.PublishLocation(context.Background(), event)

	require.NoError(t, err)
	assert.Len(t, mock.messages[TopicCourierLocation], 1)

	msg := mock.messages[TopicCourierLocation][0]
	assert.NotEmpty(t, msg.UUID)
	assert.Equal(t, "courier-1", msg.Metadata.Get("partition_key"))
	assert.Contains(t, string(msg.Payload), "courier-1")
	assert.Contains(t, string(msg.Payload), "moving")
}

func TestLocationPublisher_PublishLocationBatch(t *testing.T) {
	mock := newMockPublisher()
	publisher := &LocationPublisher{
		publisher: mock,
		logger:    watermill.NewStdLogger(false, false),
	}

	events := []vo.CourierLocationEvent{
		vo.NewCourierLocationEvent("courier-1", vo.MustNewLocation(52.5200, 13.4050), vo.CourierStatusMoving),
		vo.NewCourierLocationEvent("courier-2", vo.MustNewLocation(52.5300, 13.4150), vo.CourierStatusIdle),
		vo.NewCourierLocationEvent("courier-3", vo.MustNewLocation(52.5400, 13.4250), vo.CourierStatusDelivering),
	}

	err := publisher.PublishLocationBatch(context.Background(), events)

	require.NoError(t, err)
	assert.Len(t, mock.messages[TopicCourierLocation], 3)
}

func TestLocationPublisher_Close(t *testing.T) {
	mock := newMockPublisher()
	publisher := &LocationPublisher{
		publisher: mock,
		logger:    watermill.NewStdLogger(false, false),
	}

	err := publisher.Close()

	require.NoError(t, err)
	assert.True(t, mock.closed)
}

func TestDefaultPublisherConfig(t *testing.T) {
	config := DefaultPublisherConfig()

	assert.Equal(t, []string{"localhost:9092"}, config.Brokers)
}
