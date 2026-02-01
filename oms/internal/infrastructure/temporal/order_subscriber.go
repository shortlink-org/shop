package temporal

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/shortlink-org/go-sdk/logger"
	"go.temporal.io/sdk/client"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	queuev1 "github.com/shortlink-org/shop/oms/internal/domain/queue/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/events"
)

// OrderWorkflowName is the registered name of the order processing workflow.
const OrderWorkflowName = "OrderWorkflow"

// OrderEventSubscriber subscribes to order domain events and starts Temporal workflows.
type OrderEventSubscriber struct {
	log            logger.Logger
	temporalClient client.Client
}

// NewOrderEventSubscriber creates a new order event subscriber.
func NewOrderEventSubscriber(log logger.Logger, temporalClient client.Client) *OrderEventSubscriber {
	return &OrderEventSubscriber{
		log:            log,
		temporalClient: temporalClient,
	}
}

// Register registers the subscriber with the event publisher.
func (s *OrderEventSubscriber) Register(publisher *events.InMemoryPublisher) {
	// Subscribe to OrderCreated events
	events.SubscribeTyped(publisher, s.OnOrderCreated)

	// Subscribe to OrderCancelled events
	events.SubscribeTyped(publisher, s.OnOrderCancelled)
}

// OnOrderCreated handles OrderCreated events by starting the order workflow.
func (s *OrderEventSubscriber) OnOrderCreated(ctx context.Context, event *orderv1.OrderCreatedEvent) error {
	workflowID := fmt.Sprintf("order-%s", event.OrderID.String())

	s.log.Info("Starting order workflow",
		slog.String("workflow_id", workflowID),
		slog.String("order_id", event.OrderID.String()),
		slog.String("customer_id", event.CustomerID.String()))

	_, err := s.temporalClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                       workflowID,
		TaskQueue:                GetQueueName(queuev1.OrderTaskQueue),
		WorkflowExecutionTimeout: 24 * time.Hour, // Maximum order processing time
		// UI enrichment for better visibility in Temporal Web UI
		StaticSummary: fmt.Sprintf("Order %s for customer %s", event.OrderID.String()[:8], event.CustomerID.String()[:8]),
		StaticDetails: fmt.Sprintf(`**Order Processing Workflow**

- **Order ID:** %s
- **Customer ID:** %s
- **Items count:** %d

This workflow orchestrates the order processing saga:
1. Create order in database
2. Reserve stock
3. Process payment
4. Complete order`,
			event.OrderID.String(),
			event.CustomerID.String(),
			len(event.Items)),
	}, OrderWorkflowName, event.OrderID, event.CustomerID, event.Items)

	if err != nil {
		s.log.Error("Failed to start order workflow",
			slog.String("workflow_id", workflowID),
			slog.String("order_id", event.OrderID.String()),
			slog.Any("error", err))
		return err
	}

	return nil
}

// OnOrderCancelled handles OrderCancelled events by signaling the order workflow.
func (s *OrderEventSubscriber) OnOrderCancelled(ctx context.Context, event *orderv1.OrderCancelledEvent) error {
	workflowID := fmt.Sprintf("order-%s", event.OrderID.String())

	s.log.Info("Signaling order cancellation to workflow",
		slog.String("workflow_id", workflowID),
		slog.String("order_id", event.OrderID.String()))

	// Signal the workflow to cancel
	err := s.temporalClient.SignalWorkflow(ctx, workflowID, "", "cancel", event.Reason)
	if err != nil {
		s.log.Error("Failed to signal order workflow for cancellation",
			slog.String("workflow_id", workflowID),
			slog.String("order_id", event.OrderID.String()),
			slog.Any("error", err))
		return err
	}

	return nil
}

// Ensure OrderCreatedEvent implements ports.Event interface.
var _ ports.Event = (*orderv1.OrderCreatedEvent)(nil)

// Ensure OrderCancelledEvent implements ports.Event interface.
var _ ports.Event = (*orderv1.OrderCancelledEvent)(nil)
