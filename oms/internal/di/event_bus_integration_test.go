//go:build integration

package oms_di

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	wmkafka "github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	otelsdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/shortlink-org/go-sdk/config"
	cqrsmessage "github.com/shortlink-org/go-sdk/cqrs/message"
	"github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/go-sdk/observability/metrics"
	"github.com/testcontainers/testcontainers-go/modules/kafka"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	orderrepo "github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/testhelpers"
	uowpg "github.com/shortlink-org/shop/oms/pkg/uow/postgres"
)

func TestNewEventBus_Integration_ForwardsOutboxToKafka(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	postgres := testhelpers.SetupPostgresContainer(t)
	store, err := orderrepo.New(ctx, postgres.DB())
	require.NoError(t, err)
	t.Cleanup(store.Close)

	kafkaContainer, brokers := setupKafkaContainer(t, ctx)
	t.Cleanup(func() {
		_ = kafkaContainer.Terminate(context.Background())
	})

	cfg, err := config.New()
	require.NoError(t, err)
	cfg.Reset()
	cfg.Set("WATERMILL_KAFKA_BROKERS", brokers)

	logCfg := logger.Default()
	logCfg.Writer = io.Discard
	logCfg.Level = logger.WARN_LEVEL
	log, err := logger.New(logCfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = log.Close() })

	monitoring := &metrics.Monitoring{Metrics: otelsdkmetric.NewMeterProvider()}

	eventBus, cleanup, err := newEventBus(ctx, cfg, log, postgres.DB(), monitoring)
	require.NoError(t, err)
	t.Cleanup(cleanup)

	event := buildOrderCreatedEvent(t)
	topic := topicForEvent(event)

	subscriber, err := newKafkaSubscriber(brokers)
	require.NoError(t, err)
	t.Cleanup(func() { _ = subscriber.Close() })

	messages, err := subscriber.Subscribe(ctx, topic)
	require.NoError(t, err)

	uow := uowpg.New(postgres.Pool)
	txCtx, err := uow.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, eventBus.Publish(txCtx, event))
	require.NoError(t, uow.Commit(txCtx))

	select {
	case msg := <-messages:
		require.NotNil(t, msg)
		require.NotEmpty(t, msg.Payload)
		require.Equal(t, "event", msg.Metadata.Get(cqrsmessage.MetadataMessageKind))
		msg.Ack()
	case <-time.After(30 * time.Second):
		t.Fatalf("timeout waiting for forwarded Kafka message on topic %s", topic)
	}
}

func setupKafkaContainer(t *testing.T, ctx context.Context) (*kafka.KafkaContainer, []string) {
	t.Helper()

	container, err := kafka.Run(
		ctx,
		"confluentinc/confluent-local:7.5.0",
		kafka.WithClusterID("oms-integration-cluster"),
	)
	require.NoError(t, err)

	brokers, err := container.Brokers(ctx)
	require.NoError(t, err)

	return container, brokers
}

func newKafkaSubscriber(brokers []string) (*wmkafka.Subscriber, error) {
	saramaConfig := wmkafka.DefaultSaramaSubscriberConfig()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	return wmkafka.NewSubscriber(
		wmkafka.SubscriberConfig{
			Brokers:               brokers,
			Unmarshaler:           wmkafka.DefaultMarshaler{},
			ConsumerGroup:         "oms-event-bus-integration-" + uuid.NewString(),
			OverwriteSaramaConfig: saramaConfig,
		},
		watermill.NewStdLogger(false, false),
	)
}

func buildOrderCreatedEvent(t *testing.T) any {
	t.Helper()

	order := orderv1.NewOrderState(uuid.New())
	require.NoError(t, order.CreateOrder(context.Background(), orderv1.Items{
		orderv1.NewItem(uuid.New(), 1, decimal.NewFromFloat(9.99)),
	}))

	events := order.GetDomainEvents()
	require.Len(t, events, 1)

	return events[0]
}

func topicForEvent(event any) string {
	namer := cqrsmessage.NewShortlinkNamer("oms")

	return namer.TopicForEvent(namer.EventName(event))
}
