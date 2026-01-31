package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// LoggingCommandMiddleware wraps a CommandHandler with logging.
func LoggingCommandMiddleware[C any](log logger.Logger, next ports.CommandHandler[C]) ports.CommandHandler[C] {
	return &loggingCommandHandler[C]{log: log, next: next}
}

type loggingCommandHandler[C any] struct {
	log  logger.Logger
	next ports.CommandHandler[C]
}

func (h *loggingCommandHandler[C]) Handle(ctx context.Context, cmd C) error {
	cmdType := fmt.Sprintf("%T", cmd)
	h.log.Info("handling command", slog.String("type", cmdType))
	start := time.Now()

	err := h.next.Handle(ctx, cmd)

	duration := time.Since(start)
	if err != nil {
		h.log.Error("command failed",
			slog.String("type", cmdType),
			slog.Duration("duration", duration),
			slog.Any("error", err))
	} else {
		h.log.Info("command completed",
			slog.String("type", cmdType),
			slog.Duration("duration", duration))
	}

	return err
}

// LoggingQueryMiddleware wraps a QueryHandler with logging.
func LoggingQueryMiddleware[Q any, R any](log logger.Logger, next ports.QueryHandler[Q, R]) ports.QueryHandler[Q, R] {
	return &loggingQueryHandler[Q, R]{log: log, next: next}
}

type loggingQueryHandler[Q any, R any] struct {
	log  logger.Logger
	next ports.QueryHandler[Q, R]
}

func (h *loggingQueryHandler[Q, R]) Handle(ctx context.Context, q Q) (R, error) {
	queryType := fmt.Sprintf("%T", q)
	h.log.Info("handling query", slog.String("type", queryType))
	start := time.Now()

	result, err := h.next.Handle(ctx, q)

	duration := time.Since(start)
	if err != nil {
		h.log.Error("query failed",
			slog.String("type", queryType),
			slog.Duration("duration", duration),
			slog.Any("error", err))
	} else {
		h.log.Info("query completed",
			slog.String("type", queryType),
			slog.Duration("duration", duration))
	}

	return result, err
}

// LoggingEventMiddleware wraps an EventHandler with logging.
func LoggingEventMiddleware[E any](log logger.Logger, next ports.EventHandler[E]) ports.EventHandler[E] {
	return &loggingEventHandler[E]{log: log, next: next}
}

type loggingEventHandler[E any] struct {
	log  logger.Logger
	next ports.EventHandler[E]
}

func (h *loggingEventHandler[E]) Handle(ctx context.Context, event E) error {
	eventType := fmt.Sprintf("%T", event)
	h.log.Info("handling event", slog.String("type", eventType))
	start := time.Now()

	err := h.next.Handle(ctx, event)

	duration := time.Since(start)
	if err != nil {
		h.log.Error("event handling failed",
			slog.String("type", eventType),
			slog.Duration("duration", duration),
			slog.Any("error", err))
	} else {
		h.log.Info("event handled",
			slog.String("type", eventType),
			slog.Duration("duration", duration))
	}

	return err
}
