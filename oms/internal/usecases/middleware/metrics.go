package middleware

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

const meterName = "github.com/shortlink-org/shop/oms/usecases"

// MetricsCommandMiddleware wraps a CommandHandler with Prometheus metrics.
func MetricsCommandMiddleware[C any](next ports.CommandHandler[C]) ports.CommandHandler[C] {
	meter := otel.Meter(meterName)

	commandDuration, _ := meter.Float64Histogram(
		"oms.command.duration",
		metric.WithDescription("Duration of command execution in seconds"),
		metric.WithUnit("s"),
	)

	commandTotal, _ := meter.Int64Counter(
		"oms.command.total",
		metric.WithDescription("Total number of commands executed"),
	)

	commandErrors, _ := meter.Int64Counter(
		"oms.command.errors",
		metric.WithDescription("Total number of command errors"),
	)

	return &metricsCommandHandler[C]{
		next:            next,
		commandDuration: commandDuration,
		commandTotal:    commandTotal,
		commandErrors:   commandErrors,
	}
}

type metricsCommandHandler[C any] struct {
	next            ports.CommandHandler[C]
	commandDuration metric.Float64Histogram
	commandTotal    metric.Int64Counter
	commandErrors   metric.Int64Counter
}

func (h *metricsCommandHandler[C]) Handle(ctx context.Context, cmd C) error {
	cmdType := fmt.Sprintf("%T", cmd)
	attrs := []attribute.KeyValue{
		attribute.String("command.type", cmdType),
	}

	start := time.Now()
	err := h.next.Handle(ctx, cmd)
	duration := time.Since(start).Seconds()

	h.commandDuration.Record(ctx, duration, metric.WithAttributes(attrs...))
	h.commandTotal.Add(ctx, 1, metric.WithAttributes(attrs...))

	if err != nil {
		h.commandErrors.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	return err
}

// MetricsQueryMiddleware wraps a QueryHandler with Prometheus metrics.
func MetricsQueryMiddleware[Q any, R any](next ports.QueryHandler[Q, R]) ports.QueryHandler[Q, R] {
	meter := otel.Meter(meterName)

	queryDuration, _ := meter.Float64Histogram(
		"oms.query.duration",
		metric.WithDescription("Duration of query execution in seconds"),
		metric.WithUnit("s"),
	)

	queryTotal, _ := meter.Int64Counter(
		"oms.query.total",
		metric.WithDescription("Total number of queries executed"),
	)

	queryErrors, _ := meter.Int64Counter(
		"oms.query.errors",
		metric.WithDescription("Total number of query errors"),
	)

	return &metricsQueryHandler[Q, R]{
		next:          next,
		queryDuration: queryDuration,
		queryTotal:    queryTotal,
		queryErrors:   queryErrors,
	}
}

type metricsQueryHandler[Q any, R any] struct {
	next          ports.QueryHandler[Q, R]
	queryDuration metric.Float64Histogram
	queryTotal    metric.Int64Counter
	queryErrors   metric.Int64Counter
}

func (h *metricsQueryHandler[Q, R]) Handle(ctx context.Context, q Q) (R, error) {
	queryType := fmt.Sprintf("%T", q)
	attrs := []attribute.KeyValue{
		attribute.String("query.type", queryType),
	}

	start := time.Now()
	result, err := h.next.Handle(ctx, q)
	duration := time.Since(start).Seconds()

	h.queryDuration.Record(ctx, duration, metric.WithAttributes(attrs...))
	h.queryTotal.Add(ctx, 1, metric.WithAttributes(attrs...))

	if err != nil {
		h.queryErrors.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	return result, err
}

// MetricsEventMiddleware wraps an EventHandler with Prometheus metrics.
func MetricsEventMiddleware[E any](next ports.EventHandler[E]) ports.EventHandler[E] {
	meter := otel.Meter(meterName)

	eventDuration, _ := meter.Float64Histogram(
		"oms.event.duration",
		metric.WithDescription("Duration of event handling in seconds"),
		metric.WithUnit("s"),
	)

	eventTotal, _ := meter.Int64Counter(
		"oms.event.total",
		metric.WithDescription("Total number of events handled"),
	)

	eventErrors, _ := meter.Int64Counter(
		"oms.event.errors",
		metric.WithDescription("Total number of event handling errors"),
	)

	return &metricsEventHandler[E]{
		next:          next,
		eventDuration: eventDuration,
		eventTotal:    eventTotal,
		eventErrors:   eventErrors,
	}
}

type metricsEventHandler[E any] struct {
	next          ports.EventHandler[E]
	eventDuration metric.Float64Histogram
	eventTotal    metric.Int64Counter
	eventErrors   metric.Int64Counter
}

func (h *metricsEventHandler[E]) Handle(ctx context.Context, event E) error {
	eventType := fmt.Sprintf("%T", event)
	attrs := []attribute.KeyValue{
		attribute.String("event.type", eventType),
	}

	start := time.Now()
	err := h.next.Handle(ctx, event)
	duration := time.Since(start).Seconds()

	h.eventDuration.Record(ctx, duration, metric.WithAttributes(attrs...))
	h.eventTotal.Add(ctx, 1, metric.WithAttributes(attrs...))

	if err != nil {
		h.eventErrors.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	return err
}
