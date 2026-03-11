package oms_di

import (
	"context"
	"log/slog"

	"github.com/shortlink-org/go-sdk/config"
	"github.com/shortlink-org/go-sdk/logger"
	sdkkafka "github.com/shortlink-org/go-sdk/watermill/backends/kafka"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	omsKafka "github.com/shortlink-org/shop/oms/internal/infrastructure/kafka"
	leaderboardget "github.com/shortlink-org/shop/oms/internal/usecases/leaderboard/event/on_order_completed"
)

func NewLeaderboardConsumer(
	ctx context.Context,
	cfg *config.Config,
	log logger.Logger,
	uow ports.UnitOfWork,
	orderRepo ports.OrderRepository,
	leaderboardRepo ports.LeaderboardRepository,
) (*omsKafka.LeaderboardConsumer, func(), error) {
	cfg.SetDefault("WATERMILL_KAFKA_CONSUMER_GROUP", omsKafka.ConsumerGroupOMSLeaderboard)

	handler, err := leaderboardget.NewHandler(log, uow, orderRepo, leaderboardRepo)
	if err != nil {
		return nil, func() {}, err
	}

	subscriber, err := sdkkafka.NewSubscriberFromConfig(log, cfg)
	if err != nil {
		log.Warn("Failed to create Kafka leaderboard subscriber, running without leaderboard consumption")
		return nil, func() {}, nil //nolint:nilerr // intentionally non-fatal
	}

	consumer := omsKafka.NewLeaderboardConsumer(omsKafka.TopicOrderCompleted, subscriber, handler, log)
	if err := consumer.Start(ctx); err != nil {
		log.Warn("Failed to start Kafka leaderboard consumer", slog.Any("error", err))
		return nil, func() {}, nil //nolint:nilerr // intentionally non-fatal
	}

	cleanup := func() {
		if consumer == nil {
			return
		}
		if err := consumer.Close(); err != nil {
			log.Warn("failed to close leaderboard consumer", slog.String("error", err.Error()))
		}
	}

	return consumer, cleanup, nil
}
