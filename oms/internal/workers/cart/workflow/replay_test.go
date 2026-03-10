package cart_workflow

import (
	"io"
	"log/slog"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"

	queuev1 "github.com/shortlink-org/shop/oms/internal/domain/queue/v1"
	temporalinfra "github.com/shortlink-org/shop/oms/internal/infrastructure/temporal"
	"github.com/shortlink-org/shop/oms/internal/workers/cart/activities"
)

func TestWorkflow_ReplaySignalAndTimerResetHistory(t *testing.T) {
	t.Parallel()

	addReq := activities.AddItemRequest{
		CustomerID: testCustomerID,
		GoodID:     testGoodID,
		Quantity:   1,
		Price:      mustDecimal("9.99"),
		Discount:   mustDecimal("0"),
	}

	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflow(Workflow)

	history := &historypb.History{
		Events: []*historypb.HistoryEvent{
			cartReplayWorkflowExecutionStarted(t, 1, "Workflow", temporalinfra.GetQueueName(queuev1.CartTaskQueue), testCustomerID),
			cartReplayWorkflowTaskScheduled(2),
			cartReplayWorkflowTaskStarted(3),
			cartReplayWorkflowTaskCompleted(4, 2, 3),
			cartReplayTimerStarted(5, "5"),
			cartReplaySignalReceived(t, 6, "EVENT_ADD", addReq),
			cartReplayWorkflowTaskScheduled(7),
			cartReplayWorkflowTaskStarted(8),
			cartReplayWorkflowTaskCompleted(9, 7, 8),
			cartReplayActivityTaskScheduled(t, 10, "10", "AddItem", temporalinfra.GetQueueName(queuev1.CartTaskQueue), addReq),
			cartReplayTimerCanceled(11, "5"),
			cartReplayTimerStarted(12, "12"),
			cartReplayActivityTaskStarted(13, 10),
			cartReplayActivityTaskCompleted(14, 10, 13),
			cartReplayWorkflowTaskScheduled(15),
			cartReplayWorkflowTaskStarted(16),
		},
	}

	err := replayer.ReplayWorkflowHistory(newCartReplayLogger(), history)
	require.NoError(t, err)
}

func cartReplayWorkflowExecutionStarted(
	t *testing.T,
	eventID int64,
	workflowName string,
	taskQueue string,
	customerID interface{},
) *historypb.HistoryEvent {
	t.Helper()

	return &historypb.HistoryEvent{
		EventId:   eventID,
		EventType: enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED,
		Attributes: &historypb.HistoryEvent_WorkflowExecutionStartedEventAttributes{
			WorkflowExecutionStartedEventAttributes: &historypb.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &commonpb.WorkflowType{Name: workflowName},
				TaskQueue:    &taskqueuepb.TaskQueue{Name: taskQueue},
				Input:        cartReplayPayloads(t, customerID),
			},
		},
	}
}

func cartReplayWorkflowTaskScheduled(eventID int64) *historypb.HistoryEvent {
	return &historypb.HistoryEvent{
		EventId:   eventID,
		EventType: enumspb.EVENT_TYPE_WORKFLOW_TASK_SCHEDULED,
		Attributes: &historypb.HistoryEvent_WorkflowTaskScheduledEventAttributes{
			WorkflowTaskScheduledEventAttributes: &historypb.WorkflowTaskScheduledEventAttributes{},
		},
	}
}

func cartReplayWorkflowTaskStarted(eventID int64) *historypb.HistoryEvent {
	return &historypb.HistoryEvent{
		EventId:   eventID,
		EventType: enumspb.EVENT_TYPE_WORKFLOW_TASK_STARTED,
	}
}

func cartReplayWorkflowTaskCompleted(eventID, scheduledEventID, startedEventID int64) *historypb.HistoryEvent {
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

func cartReplaySignalReceived(t *testing.T, eventID int64, signalName string, payload interface{}) *historypb.HistoryEvent {
	t.Helper()

	return &historypb.HistoryEvent{
		EventId:   eventID,
		EventType: enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED,
		Attributes: &historypb.HistoryEvent_WorkflowExecutionSignaledEventAttributes{
			WorkflowExecutionSignaledEventAttributes: &historypb.WorkflowExecutionSignaledEventAttributes{
				SignalName: signalName,
				Input:      cartReplayPayloads(t, payload),
				Identity:   "replay-test",
			},
		},
	}
}

func cartReplayTimerStarted(eventID int64, timerID string) *historypb.HistoryEvent {
	return &historypb.HistoryEvent{
		EventId:   eventID,
		EventType: enumspb.EVENT_TYPE_TIMER_STARTED,
		Attributes: &historypb.HistoryEvent_TimerStartedEventAttributes{
			TimerStartedEventAttributes: &historypb.TimerStartedEventAttributes{
				TimerId: timerID,
			},
		},
	}
}

func cartReplayTimerCanceled(eventID int64, timerID string) *historypb.HistoryEvent {
	return &historypb.HistoryEvent{
		EventId:   eventID,
		EventType: enumspb.EVENT_TYPE_TIMER_CANCELED,
		Attributes: &historypb.HistoryEvent_TimerCanceledEventAttributes{
			TimerCanceledEventAttributes: &historypb.TimerCanceledEventAttributes{
				TimerId: timerID,
			},
		},
	}
}

func cartReplayActivityTaskScheduled(
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
				Input:        cartReplayPayloads(t, input),
			},
		},
	}
}

func cartReplayActivityTaskStarted(eventID, scheduledEventID int64) *historypb.HistoryEvent {
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

func cartReplayActivityTaskCompleted(eventID, scheduledEventID, startedEventID int64) *historypb.HistoryEvent {
	return &historypb.HistoryEvent{
		EventId:   eventID,
		EventType: enumspb.EVENT_TYPE_ACTIVITY_TASK_COMPLETED,
		Attributes: &historypb.HistoryEvent_ActivityTaskCompletedEventAttributes{
			ActivityTaskCompletedEventAttributes: &historypb.ActivityTaskCompletedEventAttributes{
				ScheduledEventId: scheduledEventID,
				StartedEventId:   startedEventID,
			},
		},
	}
}

func cartReplayPayloads(t *testing.T, values ...interface{}) *commonpb.Payloads {
	t.Helper()

	payloads, err := converter.GetDefaultDataConverter().ToPayloads(values...)
	require.NoError(t, err)

	return payloads
}

func newCartReplayLogger() log.Logger {
	return log.NewStructuredLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func mustDecimal(value string) decimal.Decimal {
	return decimal.RequireFromString(value)
}
