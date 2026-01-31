package middleware

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

const tracerName = "github.com/shortlink-org/shop/oms/usecases"

// TracingCommandMiddleware wraps a CommandHandler with OpenTelemetry tracing.
func TracingCommandMiddleware[C any](next ports.CommandHandler[C]) ports.CommandHandler[C] {
	return &tracingCommandHandler[C]{next: next, tracer: otel.Tracer(tracerName)}
}

type tracingCommandHandler[C any] struct {
	next   ports.CommandHandler[C]
	tracer trace.Tracer
}

func (h *tracingCommandHandler[C]) Handle(ctx context.Context, cmd C) error {
	cmdType := fmt.Sprintf("%T", cmd)
	ctx, span := h.tracer.Start(ctx, fmt.Sprintf("Command: %s", cmdType),
		trace.WithAttributes(
			attribute.String("command.type", cmdType),
		),
	)
	defer span.End()

	err := h.next.Handle(ctx, cmd)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return err
}

// TracingQueryMiddleware wraps a QueryHandler with OpenTelemetry tracing.
func TracingQueryMiddleware[Q any, R any](next ports.QueryHandler[Q, R]) ports.QueryHandler[Q, R] {
	return &tracingQueryHandler[Q, R]{next: next, tracer: otel.Tracer(tracerName)}
}

type tracingQueryHandler[Q any, R any] struct {
	next   ports.QueryHandler[Q, R]
	tracer trace.Tracer
}

func (h *tracingQueryHandler[Q, R]) Handle(ctx context.Context, q Q) (R, error) {
	queryType := fmt.Sprintf("%T", q)
	ctx, span := h.tracer.Start(ctx, fmt.Sprintf("Query: %s", queryType),
		trace.WithAttributes(
			attribute.String("query.type", queryType),
		),
	)
	defer span.End()

	result, err := h.next.Handle(ctx, q)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return result, err
}

// TracingEventMiddleware wraps an EventHandler with OpenTelemetry tracing.
func TracingEventMiddleware[E any](next ports.EventHandler[E]) ports.EventHandler[E] {
	return &tracingEventHandler[E]{next: next, tracer: otel.Tracer(tracerName)}
}

type tracingEventHandler[E any] struct {
	next   ports.EventHandler[E]
	tracer trace.Tracer
}

func (h *tracingEventHandler[E]) Handle(ctx context.Context, event E) error {
	eventType := fmt.Sprintf("%T", event)
	ctx, span := h.tracer.Start(ctx, fmt.Sprintf("Event: %s", eventType),
		trace.WithAttributes(
			attribute.String("event.type", eventType),
		),
	)
	defer span.End()

	err := h.next.Handle(ctx, event)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return err
}
