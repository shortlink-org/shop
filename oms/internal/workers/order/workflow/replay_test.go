package order_workflow

import (
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	queuev1 "github.com/shortlink-org/shop/oms/internal/domain/queue/v1"
	temporalinfra "github.com/shortlink-org/shop/oms/internal/infrastructure/temporal"
	"github.com/shortlink-org/shop/oms/internal/workers/order/activities"
)

func TestWorkflow_ReplayDeliveryHistory(t *testing.T) {
	t.Parallel()

	orderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	customerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174100")
	items := createTestItems()

	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(Workflow, workflow.RegisterOptions{
		Name: temporalinfra.OrderWorkflowName,
	})

	history := &historypb.History{
		Events: []*historypb.HistoryEvent{
			replayWorkflowExecutionStarted(
				t,
				1,
				temporalinfra.OrderWorkflowName,
				temporalinfra.GetQueueName(queuev1.OrderTaskQueue),
				orderID,
				customerID,
				items,
				true,
			),
			replayWorkflowTaskScheduled(2),
			replayWorkflowTaskStarted(3),
			replayWorkflowTaskCompleted(4, 2, 3),
			replayActivityTaskScheduled(
				t,
				5,
				"5",
				"RequestDelivery",
				temporalinfra.GetQueueName(queuev1.OrderTaskQueue),
				activities.RequestDeliveryRequest{OrderID: orderID},
			),
			replayActivityTaskStarted(6, 5),
			replayActivityTaskCompleted(t, 7, 5, 6, activities.RequestDeliveryResponse{
				PackageID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174999").String(),
				Status:    "ACCEPTED",
			}),
			replayWorkflowTaskScheduled(8),
			replayWorkflowTaskStarted(9),
			replayWorkflowTaskCompleted(10, 8, 9),
			replayWorkflowExecutionCompleted(11, 10),
		},
	}

	err := replayer.ReplayWorkflowHistory(newReplayLogger(), history)
	require.NoError(t, err)
}

func replayWorkflowExecutionStarted(
	t *testing.T,
	eventID int64,
	workflowName string,
	taskQueue string,
	orderID uuid.UUID,
	customerID uuid.UUID,
	items interface{},
	requestDelivery bool,
) *historypb.HistoryEvent {
	t.Helper()

	return &historypb.HistoryEvent{
		EventId:   eventID,
		EventType: enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED,
		Attributes: &historypb.HistoryEvent_WorkflowExecutionStartedEventAttributes{
			WorkflowExecutionStartedEventAttributes: &historypb.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &commonpb.WorkflowType{Name: workflowName},
				TaskQueue:    &taskqueuepb.TaskQueue{Name: taskQueue},
				Input:        replayPayloads(t, orderID, customerID, items, requestDelivery),
			},
		},
	}
}

func replayWorkflowTaskScheduled(eventID int64) *historypb.HistoryEvent {
	return &historypb.HistoryEvent{
		EventId:   eventID,
		EventType: enumspb.EVENT_TYPE_WORKFLOW_TASK_SCHEDULED,
		Attributes: &historypb.HistoryEvent_WorkflowTaskScheduledEventAttributes{
			WorkflowTaskScheduledEventAttributes: &historypb.WorkflowTaskScheduledEventAttributes{},
		},
	}
}

func replayWorkflowTaskStarted(eventID int64) *historypb.HistoryEvent {
	return &historypb.HistoryEvent{
		EventId:   eventID,
		EventType: enumspb.EVENT_TYPE_WORKFLOW_TASK_STARTED,
	}
}

func replayWorkflowTaskCompleted(eventID, scheduledEventID, startedEventID int64) *historypb.HistoryEvent {
	return &historypb.HistoryEvent{
		EventId:   eventID,
		EventType: enumspb.EVENT_TYPE_WORKFLOW_TASK_COMPLETED,
		Attributes: &historypb.HistoryEvent_WorkflowTaskCompletedEventAttributes{
			WorkflowTaskCompletedEventAttributes: &historypb.WorkflowTaskCompletedEventAttributes{
				ScheduledEventId: scheduledEventID,
				StartedEventId:   startedEventID,
			},
		},
	}
}

func replayActivityTaskScheduled(
	t *testing.T,
	eventID int64,
	activityID string,
	activityName string,
	taskQueue string,
	input interface{},
) *historypb.HistoryEvent {
	t.Helper()

	return &historypb.HistoryEvent{
		EventId:   eventID,
		EventType: enumspb.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED,
		Attributes: &historypb.HistoryEvent_ActivityTaskScheduledEventAttributes{
			ActivityTaskScheduledEventAttributes: &historypb.ActivityTaskScheduledEventAttributes{
				ActivityId:   activityID,
				ActivityType: &commonpb.ActivityType{Name: activityName},
				TaskQueue:    &taskqueuepb.TaskQueue{Name: taskQueue},
				Input:        replayPayloads(t, input),
			},
		},
	}
}

func replayActivityTaskStarted(eventID, scheduledEventID int64) *historypb.HistoryEvent {
	return &historypb.HistoryEvent{
		EventId:   eventID,
		EventType: enumspb.EVENT_TYPE_ACTIVITY_TASK_STARTED,
		Attributes: &historypb.HistoryEvent_ActivityTaskStartedEventAttributes{
			ActivityTaskStartedEventAttributes: &historypb.ActivityTaskStartedEventAttributes{
				ScheduledEventId: scheduledEventID,
			},
		},
	}
}

func replayActivityTaskCompleted(
	t *testing.T,
	eventID, scheduledEventID, startedEventID int64,
	result interface{},
) *historypb.HistoryEvent {
	t.Helper()

	return &historypb.HistoryEvent{
		EventId:   eventID,
		EventType: enumspb.EVENT_TYPE_ACTIVITY_TASK_COMPLETED,
		Attributes: &historypb.HistoryEvent_ActivityTaskCompletedEventAttributes{
			ActivityTaskCompletedEventAttributes: &historypb.ActivityTaskCompletedEventAttributes{
				ScheduledEventId: scheduledEventID,
				StartedEventId:   startedEventID,
				Result:           replayPayloads(t, result),
			},
		},
	}
}

func replayWorkflowExecutionCompleted(eventID, completedEventID int64) *historypb.HistoryEvent {
	return &historypb.HistoryEvent{
		EventId:   eventID,
		EventType: enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED,
		Attributes: &historypb.HistoryEvent_WorkflowExecutionCompletedEventAttributes{
			WorkflowExecutionCompletedEventAttributes: &historypb.WorkflowExecutionCompletedEventAttributes{
				WorkflowTaskCompletedEventId: completedEventID,
			},
		},
	}
}

func replayPayloads(t *testing.T, values ...interface{}) *commonpb.Payloads {
	t.Helper()

	payloads, err := converter.GetDefaultDataConverter().ToPayloads(values...)
	require.NoError(t, err)

	return payloads
}

func newReplayLogger() log.Logger {
	return log.NewStructuredLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))
}
