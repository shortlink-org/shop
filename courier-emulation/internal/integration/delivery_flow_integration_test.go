//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/infrastructure/kafka"
)

const (
	fullFlowConsumeTimeout   = 60 * time.Second // time for pickup + route + delivery
	fastUpdateInterval       = "100ms"
	fastPickupWait           = "300ms"
	fastDeliveryWait         = "300ms"
	simulationTimeMultiplier = "50"
	serviceStartupWait       = 8 * time.Second
)

// Berlin coordinates from plan / route_generator_test.
const (
	berlinPickupLat, berlinPickupLon     = 52.517037, 13.388860
	berlinDeliveryLat, berlinDeliveryLon = 52.529407, 13.397634
)

// TestDeliveryFlowE2E verifies the full flow: order assigned → courier moves A→B → pick_up → move to customer → deliver (DELIVERED or NOT_DELIVERED).
func TestDeliveryFlowE2E(t *testing.T) {
	kafkaC := SetupKafkaContainer(t)
	osrmC := SetupOSRMContainer(t)

	brokersStr := strings.Join(kafkaC.Brokers, ",")
	env := append(os.Environ(),
		"WATERMILL_KAFKA_BROKERS="+brokersStr,
		"OSRM_URL="+osrmC.BaseURL,
		"SIMULATION_UPDATE_INTERVAL="+fastUpdateInterval,
		"SIMULATION_PICKUP_WAIT="+fastPickupWait,
		"SIMULATION_DELIVERY_WAIT="+fastDeliveryWait,
		"SIMULATION_TIME_MULTIPLIER="+simulationTimeMultiplier,
	)

	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "courier-emulation")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/courier-emulation")
	buildCmd.Dir = repoRoot(t)
	buildCmd.Env = os.Environ()
	out, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "build failed: %s", string(out))

	cmd := exec.Command(binPath)
	cmd.Dir = repoRoot(t)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Start())
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Signal(os.Interrupt)
			_ = cmd.Wait()
		}
	}()

	time.Sleep(serviceStartupWait)

	event := kafka.OrderAssignedEvent{
		PackageID:       "pkg-e2e-1",
		CourierID:       "courier-e2e-1",
		AssignedAt:      time.Now().Add(-time.Minute),
		PickupAddress:   kafka.Address{Latitude: berlinPickupLat, Longitude: berlinPickupLon},
		DeliveryAddress: kafka.Address{Latitude: berlinDeliveryLat, Longitude: berlinDeliveryLon},
	}
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	producer, err := sarama.NewSyncProducer(kafkaC.Brokers, cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = producer.Close() })

	var (
		locations  []locationMsg
		pickUps    []pickUpMsg
		deliveries []deliverOrderMsg
		mu         sync.Mutex
	)
	multiHandler := &multiTopicHandler{
		onMessage: func(topic string, b []byte) {
			mu.Lock()
			defer mu.Unlock()
			switch topic {
			case kafka.TopicCourierLocation:
				var m locationMsg
				if json.Unmarshal(b, &m) == nil {
					locations = append(locations, m)
				}
			case kafka.TopicPickUpOrder:
				var m pickUpMsg
				if json.Unmarshal(b, &m) == nil {
					pickUps = append(pickUps, m)
				}
			case kafka.TopicDeliverOrder:
				var m deliverOrderMsg
				if json.Unmarshal(b, &m) == nil {
					deliveries = append(deliveries, m)
				}
			}
		},
	}

	consumer, err := sarama.NewConsumerGroup(kafkaC.Brokers, "integration-e2e", cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = consumer.Close() })

	topics := []string{kafka.TopicCourierLocation, kafka.TopicPickUpOrder, kafka.TopicDeliverOrder}
	consumeCtx, consumeCancel := context.WithTimeout(context.Background(), fullFlowConsumeTimeout)
	t.Cleanup(consumeCancel)

	// Start consumer before producing so we don't miss pick_up or early location events
	go func() {
		for {
			if err := consumer.Consume(consumeCtx, topics, multiHandler); err != nil {
				return
			}
			if consumeCtx.Err() != nil {
				return
			}
		}
	}()
	time.Sleep(2 * time.Second) // let consumer join and get partition assignments

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: kafka.TopicOrderAssigned,
		Value: sarama.ByteEncoder(payload),
	})
	require.NoError(t, err)

	// Wait until we get at least one delivery outcome (DELIVERED or NOT_DELIVERED)
	deadline := time.Now().Add(fullFlowConsumeTimeout)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(deliveries)
		mu.Unlock()
		if n >= 1 {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	mu.Lock()
	locCount := len(locations)
	pickUpCount := len(pickUps)
	deliverCount := len(deliveries)
	deliveriesCopy := make([]deliverOrderMsg, len(deliveries))
	copy(deliveriesCopy, deliveries)
	mu.Unlock()

	require.GreaterOrEqual(t, locCount, 1, "expected at least one location update (courier moved)")
	require.Equal(t, 1, pickUpCount, "expected exactly one pick_up_order event")
	require.GreaterOrEqual(t, deliverCount, 1, "expected at least one deliver_order event (DELIVERED or NOT_DELIVERED)")

	deliver := deliveriesCopy[0]
	assert.Equal(t, event.CourierID, deliver.CourierID, "deliver_order.courier_id should match")
	assert.Equal(t, event.PackageID, deliver.PackageID, "deliver_order.package_id should match")
	assert.Contains(t, []string{kafka.DeliveryStatusDelivered, kafka.DeliveryStatusNotDelivered}, deliver.Status,
		"deliver_order.status must be DELIVERED or NOT_DELIVERED, got %q", deliver.Status)
	if deliver.Status == kafka.DeliveryStatusNotDelivered {
		assert.NotEmpty(t, deliver.Reason, "NOT_DELIVERED should have reason")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	// Walk up to find go.mod (module root).
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

type locationMsg struct {
	CourierID string  `json:"courier_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Status    string  `json:"status"`
}

type pickUpMsg struct {
	PackageID  string `json:"package_id"`
	CourierID  string `json:"courier_id"`
	PickedUpAt string `json:"picked_up_at"`
}

type deliverOrderMsg struct {
	PackageID   string `json:"package_id"`
	CourierID   string `json:"courier_id"`
	Status      string `json:"status"`
	Reason      string `json:"reason,omitempty"`
	DeliveredAt string `json:"delivered_at"`
}

// multiTopicHandler dispatches messages to onMessage with topic name.
type multiTopicHandler struct {
	onMessage func(topic string, payload []byte)
}

func (h *multiTopicHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *multiTopicHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *multiTopicHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	topic := claim.Topic()
	for msg := range claim.Messages() {
		if msg != nil && msg.Value != nil && h.onMessage != nil {
			h.onMessage(topic, msg.Value)
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
