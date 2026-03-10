package oms_di

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	wmkafka "github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	wmsql "github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shortlink-org/go-sdk/config"
	"github.com/shortlink-org/go-sdk/cqrs/bus"
	cqrsmessage "github.com/shortlink-org/go-sdk/cqrs/message"
	"github.com/shortlink-org/go-sdk/db"
	logger "github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/go-sdk/observability/metrics"
	sdkwatermill "github.com/shortlink-org/go-sdk/watermill"
)

const (
	outboxForwarderTopic   = "oms_outbox"
	outboxConsumerGroup    = "oms-outbox-forwarder"
	defaultOutboxPollDelay = 100 * time.Millisecond
	defaultAckDeadline     = 5 * time.Second
	defaultForwarderClose  = 5 * time.Second
)

// newEventBus creates EventBus with tx-aware outbox and a running forwarder.
func newEventBus(
	ctx context.Context,
	cfg *config.Config,
	log logger.Logger,
	store db.DB,
	monitoring *metrics.Monitoring,
) (*bus.EventBus, func(), error) {
	pool, ok := store.GetConn().(*pgxpool.Pool)
	if !ok {
		return nil, nil, db.ErrGetConnection
	}

	brokers := cfg.GetStringSlice("WATERMILL_KAFKA_BROKERS")
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}

	wmLogger := sdkwatermill.NewWatermillLogger(log)
	namer := cqrsmessage.NewShortlinkNamer("oms")
	marshaler := cqrsmessage.NewJSONMarshaler(namer)

	realPublisher, err := wmkafka.NewPublisher(wmkafka.PublisherConfig{
		Brokers: brokers,
	}, wmLogger)
	if err != nil {
		return nil, nil, fmt.Errorf("create kafka outbox publisher: %w", err)
	}

	sqlSubscriber, err := wmsql.NewSubscriber(
		wmsql.BeginnerFromPgx(pool),
		wmsql.SubscriberConfig{
			SchemaAdapter:    wmsql.DefaultPostgreSQLSchema{},
			OffsetsAdapter:   wmsql.DefaultPostgreSQLOffsetsAdapter{},
			InitializeSchema: true,
			ConsumerGroup:    outboxConsumerGroup,
			PollInterval:     defaultOutboxPollDelay,
			AckDeadline:      ptrDuration(defaultAckDeadline),
		},
		wmLogger,
	)
	if err != nil {
		_ = realPublisher.Close()
		return nil, nil, fmt.Errorf("create sql outbox subscriber: %w", err)
	}

	if err := sqlSubscriber.SubscribeInitialize(outboxForwarderTopic); err != nil {
		_ = sqlSubscriber.Close()
		_ = realPublisher.Close()
		return nil, nil, fmt.Errorf("initialize outbox topic: %w", err)
	}

	eventBus, err := bus.NewEventBusWithOptions(
		realPublisher,
		marshaler,
		namer,
		bus.WithTxAwareOutbox(outboxForwarderTopic, wmLogger),
		bus.WithOutbox(bus.OutboxConfig{
			Pool:          pool,
			Subscriber:    sqlSubscriber,
			RealPublisher: realPublisher,
			ForwarderName: outboxForwarderTopic,
			Logger:        log,
			MeterProvider: monitoring.Metrics,
		}),
	)
	if err != nil {
		_ = sqlSubscriber.Close()
		_ = realPublisher.Close()
		return nil, nil, fmt.Errorf("create event bus: %w", err)
	}

	forwarderCtx, stopForwarder := context.WithCancel(ctx)
	go func() {
		if runErr := eventBus.RunForwarder(forwarderCtx); runErr != nil && forwarderCtx.Err() == nil {
			log.Error("CQRS outbox forwarder stopped with error", slog.String("error", runErr.Error()))
		}
	}()

	cleanup := func() {
		stopForwarder()

		closeCtx, cancel := context.WithTimeout(context.Background(), defaultForwarderClose)
		defer cancel()

		if err := eventBus.CloseForwarder(closeCtx); err != nil {
			log.Warn("failed to close event bus forwarder", slog.String("error", err.Error()))
		}

		if err := sqlSubscriber.Close(); err != nil {
			log.Warn("failed to close sql outbox subscriber", slog.String("error", err.Error()))
		}

		if err := realPublisher.Close(); err != nil {
			log.Warn("failed to close kafka outbox publisher", slog.String("error", err.Error()))
		}
	}

	return eventBus, cleanup, nil
}

func ptrDuration(value time.Duration) *time.Duration {
	return &value
}
