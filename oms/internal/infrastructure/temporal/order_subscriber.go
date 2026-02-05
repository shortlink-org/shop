package temporal

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/shortlink-org/go-sdk/logger"
	"go.temporal.io/sdk/client"

	eventsv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/events/v1"
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
	orderRepo      ports.OrderRepository
}

// NewOrderEventSubscriber creates a new order event subscriber.
func NewOrderEventSubscriber(log logger.Logger, temporalClient client.Client, orderRepo ports.OrderRepository) *OrderEventSubscriber {
	return &OrderEventSubscriber{
		log:            log,
		temporalClient: temporalClient,
		orderRepo:      orderRepo,
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
// It loads the order to get items and whether delivery was requested, then starts the workflow.
func (s *OrderEventSubscriber) OnOrderCreated(ctx context.Context, event *eventsv1.OrderCreated) error {
	orderID, err := uuid.Parse(event.GetOrderId())
	if err != nil {
		return err
	}
	customerID, err := uuid.Parse(event.GetCustomerId())
	if err != nil {
		return err
	}

	// Load order to get items (for workflow) and whether delivery was requested
	order, err := s.orderRepo.Load(ctx, orderID)
	if err != nil {
		s.log.Error("Failed to load order for workflow",
			slog.String("order_id", orderID.String()),
			slog.Any("error", err))
		return err
	}

	items := order.GetItems()
	requestDelivery := order.HasDeliveryInfo()

	workflowID := fmt.Sprintf("order-%s", orderID.String())

	s.log.Info("Starting order workflow",
		slog.String("workflow_id", workflowID),
		slog.String("order_id", orderID.String()),
		slog.String("customer_id", customerID.String()),
		slog.Bool("request_delivery", requestDelivery))

	_, err = s.temporalClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                       workflowID,
		TaskQueue:                GetQueueName(queuev1.OrderTaskQueue),
		WorkflowExecutionTimeout: 24 * time.Hour, // Maximum order processing time
		StaticSummary:            fmt.Sprintf("Order %s for customer %s", orderID.String()[:8], customerID.String()[:8]),
		StaticDetails: fmt.Sprintf(`**Order Processing Workflow**

- **Order ID:** %s
- **Customer ID:** %s
- **Items count:** %d
- **Request delivery:** %t`,
			orderID.String(),
			customerID.String(),
			len(items),
			requestDelivery),
	}, OrderWorkflowName, orderID, customerID, items, requestDelivery)

	if err != nil {
		s.log.Error("Failed to start order workflow",
			slog.String("workflow_id", workflowID),
			slog.String("order_id", orderID.String()),
			slog.Any("error", err))
		return err
	}

	return nil
}

// OnOrderCancelled handles OrderCancelled events by signaling the order workflow.
func (s *OrderEventSubscriber) OnOrderCancelled(ctx context.Context, event *eventsv1.OrderCancelled) error {
	workflowID := fmt.Sprintf("order-%s", event.GetOrderId())

	s.log.Info("Signaling order cancellation to workflow",
		slog.String("workflow_id", workflowID),
		slog.String("order_id", event.GetOrderId()))

	// Signal the workflow to cancel
	err := s.temporalClient.SignalWorkflow(ctx, workflowID, "", "cancel", event.GetReason())
	if err != nil {
		s.log.Error("Failed to signal order workflow for cancellation",
			slog.String("workflow_id", workflowID),
			slog.String("order_id", event.GetOrderId()),
			slog.Any("error", err))
		return err
	}

	return nil
}
