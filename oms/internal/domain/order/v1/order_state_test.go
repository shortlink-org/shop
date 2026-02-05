package v1

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/shortlink-org/go-sdk/fsm"
	"github.com/stretchr/testify/require"

	common "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common"
	"github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/address"
)

func TestOrderState(t *testing.T) {
	// Define fixed UUIDs for consistency across tests.
	fixedCustomerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedGoodID1 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")
	fixedGoodID2 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174002")

	t.Run("NewOrderState", func(t *testing.T) {
		orderState := NewOrderState(fixedCustomerID)

		require.Equal(t, fixedCustomerID, orderState.GetCustomerId(), "Customer ID should match")
		require.Equal(t, OrderStatus_ORDER_STATUS_PENDING, orderState.GetStatus(), "Initial status should be Pending")
		require.Empty(t, orderState.GetItems(), "Initial items should be empty")
	})

	t.Run("CreateOrder", func(t *testing.T) {
		orderState := NewOrderState(fixedCustomerID)

		items := Items{
			NewItem(fixedGoodID1, 2, decimal.NewFromFloat(19.99)),
			NewItem(fixedGoodID2, 1, decimal.NewFromFloat(9.99)),
		}

		err := orderState.CreateOrder(context.Background(), items)
		require.NoError(t, err, "CreateOrder should not return an error")
		require.Equal(t, OrderStatus_ORDER_STATUS_PROCESSING, orderState.GetStatus(), "Status should transition to Processing")
		require.Equal(t, items, orderState.GetItems(), "Items should match the created items")
	})

	t.Run("UpdateOrder", func(t *testing.T) {
		orderState := NewOrderState(fixedCustomerID)

		initialItems := Items{
			NewItem(fixedGoodID1, 2, decimal.NewFromFloat(19.99)),
		}
		err := orderState.CreateOrder(context.Background(), initialItems)
		require.NoError(t, err, "CreateOrder should not return an error")

		updatedItems := Items{
			NewItem(fixedGoodID1, 3, decimal.NewFromFloat(29.99)),
			NewItem(fixedGoodID2, 1, decimal.NewFromFloat(9.99)),
		}

		err = orderState.UpdateOrder(updatedItems)
		require.NoError(t, err, "UpdateOrder should not return an error")
		require.Equal(t, updatedItems, orderState.GetItems(), "Items should reflect the updates")
	})

	t.Run("CancelOrder", func(t *testing.T) {
		orderState := NewOrderState(fixedCustomerID)

		items := Items{
			NewItem(fixedGoodID1, 2, decimal.NewFromFloat(19.99)),
		}
		err := orderState.CreateOrder(context.Background(), items)
		require.NoError(t, err, "CreateOrder should not return an error")

		err = orderState.CancelOrder()
		require.NoError(t, err, "CancelOrder should not return an error")
		require.Equal(t, OrderStatus_ORDER_STATUS_CANCELLED, orderState.GetStatus(), "Status should transition to Canceled")
	})

	t.Run("CompleteOrder", func(t *testing.T) {
		orderState := NewOrderState(fixedCustomerID)

		items := Items{
			NewItem(fixedGoodID1, 2, decimal.NewFromFloat(19.99)),
		}
		err := orderState.CreateOrder(context.Background(), items)
		require.NoError(t, err, "CreateOrder should not return an error")

		err = orderState.CompleteOrder()
		require.NoError(t, err, "CompleteOrder should not return an error")
		require.Equal(t, OrderStatus_ORDER_STATUS_COMPLETED, orderState.GetStatus(), "Status should transition to Completed")
	})

	t.Run("OrderStateConcurrency", func(t *testing.T) {
		orderState := NewOrderState(fixedCustomerID)

		items := Items{
			NewItem(fixedGoodID1, 2, decimal.NewFromFloat(19.99)),
			NewItem(fixedGoodID2, 1, decimal.NewFromFloat(9.99)),
		}

		err := orderState.CreateOrder(context.Background(), items)
		require.NoError(t, err, "CreateOrder should not return an error")

		updatedItems := Items{
			NewItem(fixedGoodID1, 3, decimal.NewFromFloat(29.99)),
			NewItem(fixedGoodID2, 2, decimal.NewFromFloat(19.99)),
		}

		var wg sync.WaitGroup
		wg.Add(2)

		// Simulate concurrent operations
		go func() {
			defer wg.Done()

			err := orderState.UpdateOrder(updatedItems)
			// It's possible that UpdateOrder happens before or after CancelOrder.
			// Depending on the FSM, updating after cancellation might fail or be allowed.
			// Here, we assume it succeeds or fails gracefully.
			if err != nil {
				t.Logf("UpdateOrder encountered an error: %v", err)
			}
		}()

		go func() {
			defer wg.Done()

			err := orderState.CancelOrder()
			if err != nil {
				t.Logf("CancelOrder encountered an error: %v", err)
			}
		}()

		wg.Wait()

		// After concurrent operations, the order should either be Canceled or have updated items.
		// Depending on the FSM's transition rules, updating after cancellation might not change the state.
		finalStatus := orderState.GetStatus()
		require.True(t, finalStatus == OrderStatus_ORDER_STATUS_CANCELLED || finalStatus == OrderStatus_ORDER_STATUS_PROCESSING,
			"Final status should be either Canceled or Processing")
	})

	t.Run("Callbacks", func(t *testing.T) {
		orderState := NewOrderState(fixedCustomerID)

		// Variables to track callbacks.
		var (
			enterState   string
			exitState    string
			triggeredEvt fsm.Event
			callbackMu   sync.Mutex
		)

		// Set up the OnEnterState callback.
		orderState.fsm.SetOnEnterState(func(ctx context.Context, from, to fsm.State, event fsm.Event) {
			callbackMu.Lock()
			defer callbackMu.Unlock()

			enterState = to.String()
			triggeredEvt = event
		})

		// Set up the OnExitState callback.
		orderState.fsm.SetOnExitState(func(ctx context.Context, from, to fsm.State, event fsm.Event) {
			callbackMu.Lock()
			defer callbackMu.Unlock()

			exitState = from.String()
			triggeredEvt = event
		})

		// Transition: Pending -> Processing
		err := orderState.CreateOrder(context.Background(), Items{
			NewItem(fixedGoodID1, 2, decimal.NewFromFloat(19.99)),
		})
		require.NoError(t, err, "CreateOrder should transition state to Processing")

		// Verify callbacks.
		callbackMu.Lock()
		require.Equal(t, "ORDER_STATUS_PENDING", exitState, "OnExitState should be called with Pending")
		require.Equal(t, fsm.Event("ORDER_TRANSITION_EVENT_CREATE"), triggeredEvt, "Triggered event should be ORDER_TRANSITION_EVENT_CREATE")
		require.Equal(t, "ORDER_STATUS_PROCESSING", enterState, "OnEnterState should be called with Processing")
		callbackMu.Unlock()

		// Reset callback trackers.
		callbackMu.Lock()

		enterState, exitState, triggeredEvt = "", "", ""

		callbackMu.Unlock()

		// Transition: Processing -> Completed
		err = orderState.CompleteOrder()
		require.NoError(t, err, "CompleteOrder should transition state to Completed")

		// Verify callbacks.
		callbackMu.Lock()
		require.Equal(t, "ORDER_STATUS_PROCESSING", exitState, "OnExitState should be called with Processing")
		require.Equal(t, fsm.Event("ORDER_TRANSITION_EVENT_COMPLETE"), triggeredEvt, "Triggered event should be ORDER_TRANSITION_EVENT_COMPLETE")
		require.Equal(t, "ORDER_STATUS_COMPLETED", enterState, "OnEnterState should be called with Completed")
		callbackMu.Unlock()
	})

	t.Run("InvalidTransitions", func(t *testing.T) {
		orderState := NewOrderState(fixedCustomerID)

		// Attempt to cancel the order while it's in Pending state.
		err := orderState.CancelOrder()
		require.NoError(t, err, "CancelOrder should transition state to Canceled from Pending")

		// Attempt to complete a Canceled order.
		err = orderState.CompleteOrder()
		require.Error(t, err, "CompleteOrder should return an error when transitioning from Canceled")
		require.Equal(t, OrderStatus_ORDER_STATUS_CANCELLED, orderState.GetStatus(), "Status should remain Canceled after invalid transition")
	})

	// ContextCancellation test removed: domain layer no longer depends on context.Context
	// Domain methods use context.Background() internally for FSM, keeping domain pure
}

// createTestDeliveryInfo creates a test DeliveryInfo for testing purposes.
func createTestDeliveryInfo(t *testing.T) DeliveryInfo {
	t.Helper()

	pickupAddr, err := address.NewAddress("123 Warehouse St", "Moscow", "101000", "Russia")
	require.NoError(t, err)

	deliveryAddr, err := address.NewAddress("456 Customer St", "Moscow", "102000", "Russia")
	require.NoError(t, err)

	// Create a valid future delivery period
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	period := NewDeliveryPeriod(startTime, endTime)

	packageInfo := NewPackageInfo(2.5)

	return NewDeliveryInfo(pickupAddr, deliveryAddr, period, packageInfo, DeliveryPriorityNormal, nil)
}

func TestSetDeliveryInfo_OrderStatusValidation(t *testing.T) {
	fixedCustomerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedGoodID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")

	t.Run("AllowsSettingDeliveryInfoInPendingState", func(t *testing.T) {
		order := NewOrderState(fixedCustomerID)
		deliveryInfo := createTestDeliveryInfo(t)

		err := order.SetDeliveryInfo(deliveryInfo)
		require.NoError(t, err, "SetDeliveryInfo should succeed in PENDING state")
		require.True(t, order.HasDeliveryInfo(), "Order should have delivery info")
	})

	t.Run("AllowsSettingDeliveryInfoInProcessingState", func(t *testing.T) {
		order := NewOrderState(fixedCustomerID)
		items := Items{NewItem(fixedGoodID, 1, decimal.NewFromFloat(10.00))}
		err := order.CreateOrder(context.Background(), items)
		require.NoError(t, err)

		deliveryInfo := createTestDeliveryInfo(t)
		err = order.SetDeliveryInfo(deliveryInfo)
		require.NoError(t, err, "SetDeliveryInfo should succeed in PROCESSING state")
	})

	t.Run("BlocksSettingDeliveryInfoInCompletedState", func(t *testing.T) {
		order := NewOrderState(fixedCustomerID)
		items := Items{NewItem(fixedGoodID, 1, decimal.NewFromFloat(10.00))}
		err := order.CreateOrder(context.Background(), items)
		require.NoError(t, err)
		err = order.CompleteOrder()
		require.NoError(t, err)

		deliveryInfo := createTestDeliveryInfo(t)
		err = order.SetDeliveryInfo(deliveryInfo)
		require.Error(t, err, "SetDeliveryInfo should fail in COMPLETED state")
		require.Contains(t, err.Error(), "ORDER_STATUS_COMPLETED")
	})

	t.Run("BlocksSettingDeliveryInfoInCancelledState", func(t *testing.T) {
		order := NewOrderState(fixedCustomerID)
		err := order.CancelOrder()
		require.NoError(t, err)

		deliveryInfo := createTestDeliveryInfo(t)
		err = order.SetDeliveryInfo(deliveryInfo)
		require.Error(t, err, "SetDeliveryInfo should fail in CANCELED state")
		require.Contains(t, err.Error(), "ORDER_STATUS_CANCELLED")
	})
}

func TestSetDeliveryInfo_DeliveryStatusValidation(t *testing.T) {
	fixedCustomerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedGoodID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")

	t.Run("AllowsSettingDeliveryInfoWhenDeliveryStatusAccepted", func(t *testing.T) {
		order := NewOrderState(fixedCustomerID)
		items := Items{NewItem(fixedGoodID, 1, decimal.NewFromFloat(10.00))}
		err := order.CreateOrder(context.Background(), items)
		require.NoError(t, err)

		// Set delivery status to ACCEPTED
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_ACCEPTED)
		require.NoError(t, err)

		deliveryInfo := createTestDeliveryInfo(t)
		err = order.SetDeliveryInfo(deliveryInfo)
		require.NoError(t, err, "SetDeliveryInfo should succeed when delivery is ACCEPTED")
	})

	t.Run("BlocksSettingDeliveryInfoWhenDeliveryStatusAssigned", func(t *testing.T) {
		order := NewOrderState(fixedCustomerID)
		items := Items{NewItem(fixedGoodID, 1, decimal.NewFromFloat(10.00))}
		err := order.CreateOrder(context.Background(), items)
		require.NoError(t, err)

		// Set delivery status to ASSIGNED (courier assigned)
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_ACCEPTED)
		require.NoError(t, err)
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_ASSIGNED)
		require.NoError(t, err)

		deliveryInfo := createTestDeliveryInfo(t)
		err = order.SetDeliveryInfo(deliveryInfo)
		require.Error(t, err, "SetDeliveryInfo should fail when courier is ASSIGNED")
		require.Contains(t, err.Error(), "DELIVERY_STATUS_ASSIGNED")
	})

	t.Run("BlocksSettingDeliveryInfoWhenDeliveryStatusInTransit", func(t *testing.T) {
		order := NewOrderState(fixedCustomerID)
		items := Items{NewItem(fixedGoodID, 1, decimal.NewFromFloat(10.00))}
		err := order.CreateOrder(context.Background(), items)
		require.NoError(t, err)

		// Set delivery status to IN_TRANSIT
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_ACCEPTED)
		require.NoError(t, err)
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_ASSIGNED)
		require.NoError(t, err)
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT)
		require.NoError(t, err)

		deliveryInfo := createTestDeliveryInfo(t)
		err = order.SetDeliveryInfo(deliveryInfo)
		require.Error(t, err, "SetDeliveryInfo should fail when package is IN_TRANSIT")
		require.Contains(t, err.Error(), "DELIVERY_STATUS_IN_TRANSIT")
	})

	t.Run("BlocksSettingDeliveryInfoWhenDeliveryStatusDelivered", func(t *testing.T) {
		order := NewOrderState(fixedCustomerID)
		items := Items{NewItem(fixedGoodID, 1, decimal.NewFromFloat(10.00))}
		err := order.CreateOrder(context.Background(), items)
		require.NoError(t, err)

		// Set delivery status to DELIVERED
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_ACCEPTED)
		require.NoError(t, err)
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_ASSIGNED)
		require.NoError(t, err)
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT)
		require.NoError(t, err)
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_DELIVERED)
		require.NoError(t, err)

		deliveryInfo := createTestDeliveryInfo(t)
		err = order.SetDeliveryInfo(deliveryInfo)
		require.Error(t, err, "SetDeliveryInfo should fail when package is DELIVERED")
		require.Contains(t, err.Error(), "DELIVERY_STATUS_DELIVERED")
	})
}

func TestSetDeliveryStatus(t *testing.T) {
	fixedCustomerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedGoodID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")

	t.Run("AllowsValidDeliveryStatusTransitions", func(t *testing.T) {
		order := NewOrderState(fixedCustomerID)
		items := Items{NewItem(fixedGoodID, 1, decimal.NewFromFloat(10.00))}
		err := order.CreateOrder(context.Background(), items)
		require.NoError(t, err)

		// UNSPECIFIED -> ACCEPTED
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_ACCEPTED)
		require.NoError(t, err)
		require.Equal(t, common.DeliveryStatus_DELIVERY_STATUS_ACCEPTED, order.GetDeliveryStatus())

		// ACCEPTED -> ASSIGNED
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_ASSIGNED)
		require.NoError(t, err)
		require.Equal(t, common.DeliveryStatus_DELIVERY_STATUS_ASSIGNED, order.GetDeliveryStatus())

		// ASSIGNED -> IN_TRANSIT
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT)
		require.NoError(t, err)
		require.Equal(t, common.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT, order.GetDeliveryStatus())

		// IN_TRANSIT -> DELIVERED
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_DELIVERED)
		require.NoError(t, err)
		require.Equal(t, common.DeliveryStatus_DELIVERY_STATUS_DELIVERED, order.GetDeliveryStatus())
	})

	t.Run("AllowsNotDeliveredFromInTransit", func(t *testing.T) {
		order := NewOrderState(fixedCustomerID)
		items := Items{NewItem(fixedGoodID, 1, decimal.NewFromFloat(10.00))}
		err := order.CreateOrder(context.Background(), items)
		require.NoError(t, err)

		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_ACCEPTED)
		require.NoError(t, err)
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_ASSIGNED)
		require.NoError(t, err)
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT)
		require.NoError(t, err)

		// IN_TRANSIT -> NOT_DELIVERED
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED)
		require.NoError(t, err)
		require.Equal(t, common.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED, order.GetDeliveryStatus())
	})

	t.Run("BlocksInvalidDeliveryStatusTransition", func(t *testing.T) {
		order := NewOrderState(fixedCustomerID)
		items := Items{NewItem(fixedGoodID, 1, decimal.NewFromFloat(10.00))}
		err := order.CreateOrder(context.Background(), items)
		require.NoError(t, err)

		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_ACCEPTED)
		require.NoError(t, err)

		// ACCEPTED -> DELIVERED (invalid - should go through ASSIGNED and IN_TRANSIT)
		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_DELIVERED)
		require.Error(t, err, "Should not allow skipping delivery status transitions")
		require.Contains(t, err.Error(), "invalid delivery status transition")
	})

	t.Run("BlocksDeliveryStatusUpdateInCompletedOrder", func(t *testing.T) {
		order := NewOrderState(fixedCustomerID)
		items := Items{NewItem(fixedGoodID, 1, decimal.NewFromFloat(10.00))}
		err := order.CreateOrder(context.Background(), items)
		require.NoError(t, err)
		err = order.CompleteOrder()
		require.NoError(t, err)

		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_ACCEPTED)
		require.Error(t, err, "Should not allow delivery status update in COMPLETED order")
		require.Contains(t, err.Error(), "ORDER_STATUS_COMPLETED")
	})

	t.Run("BlocksDeliveryStatusUpdateInCancelledOrder", func(t *testing.T) {
		order := NewOrderState(fixedCustomerID)
		err := order.CancelOrder()
		require.NoError(t, err)

		err = order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_ACCEPTED)
		require.Error(t, err, "Should not allow delivery status update in CANCELED order")
		require.Contains(t, err.Error(), "ORDER_STATUS_CANCELLED")
	})
}
