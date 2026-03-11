package kafka

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	cqrsmessage "github.com/shortlink-org/go-sdk/cqrs/message"
	logger "github.com/shortlink-org/go-sdk/logger"

	orderevents "github.com/shortlink-org/shop/oms/internal/domain/order/v1/events/v1"
)

const (
	ConsumerGroupOMSLeaderboard = "oms-leaderboard-consumer"
	TopicOrderCompleted         = "oms.order.completed.v1"
)

type OrderCompletedHandler interface {
	Handle(ctx context.Context, event *orderevents.OrderCompleted) error
}

type LeaderboardConsumer struct {
	topic      string
	subscriber message.Subscriber
	handler    OrderCompletedHandler
	log        logger.Logger
	marshaler  *cqrsmessage.JSONMarshaler
	cancel     context.CancelCauseFunc
}

func NewLeaderboardConsumer(
	topic string,
	subscriber message.Subscriber,
	handler OrderCompletedHandler,
	log logger.Logger,
) *LeaderboardConsumer {
	return &LeaderboardConsumer{
		topic:      topic,
		subscriber: subscriber,
		handler:    handler,
		log:        log,
		marshaler:  cqrsmessage.NewJSONMarshaler(cqrsmessage.NewShortlinkNamer("oms")),
	}
}

func (c *LeaderboardConsumer) Start(ctx context.Context) error {
	c.log.Info("Starting leaderboard consumer", slog.String("topic", c.topic))

	messages, err := c.subscriber.Subscribe(ctx, c.topic)
	if err != nil {
		return fmt.Errorf("subscribe leaderboard topic: %w", err)
	}

	ctx, c.cancel = context.WithCancelCause(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-messages:
				if msg == nil {
					continue
				}
				c.processMessage(ctx, msg)
			}
		}
	}()

	return nil
}

func (c *LeaderboardConsumer) processMessage(ctx context.Context, msg *message.Message) {
	var event orderevents.OrderCompleted

	if err := c.marshaler.Unmarshal(msg, &event); err != nil {
		c.log.Error("failed to decode completed-order event for leaderboard",
			slog.String("uuid", msg.UUID),
			slog.String("error", err.Error()))
		msg.Ack()

		return
	}

	if err := c.handler.Handle(ctx, &event); err != nil {
		c.log.Error("failed to apply completed-order leaderboard projection",
			slog.String("uuid", msg.UUID),
			slog.String("order_id", event.GetOrderId()),
			slog.String("error", err.Error()))
		msg.Nack()

		return
	}

	msg.Ack()
}

func (c *LeaderboardConsumer) Close() error {
	if c.cancel != nil {
		c.cancel(fmt.Errorf("leaderboard consumer closed"))
	}

	if err := c.subscriber.Close(); err != nil {
		return fmt.Errorf("close leaderboard subscriber: %w", err)
	}

	return nil
}
