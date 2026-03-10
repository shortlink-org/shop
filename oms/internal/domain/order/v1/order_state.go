package v1 //nolint:funlen,funcorder // FSM and helpers; order kept for readability

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shortlink-org/go-sdk/fsm"
	"google.golang.org/protobuf/types/known/timestamppb"

	domainevents "github.com/shortlink-org/shop/oms/internal/domain/events"
	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common"
	eventsv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/events/v1"
	addressvo "github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/address"
)

// OrderState represents the order state.
type OrderState struct {
	mu sync.Mutex

	// orderID is the order ID
	id uuid.UUID
	// items is the list of order items
	items Items
	// customerId is the customer ID
	customerId uuid.UUID
	// version is used for optimistic concurrency control
	version int
	// fsm is the finite state machine for the order status
	fsm *fsm.FSM
	// domainEvents stores domain events (proto) that occurred during aggregate operations
	domainEvents []domainevents.Event
	// deliveryInfo contains delivery information for the order (nil = self-pickup)
	deliveryInfo *DeliveryInfo
	// deliveryStatus tracks the delivery status (ACCEPTED, ASSIGNED, IN_TRANSIT, etc.)
	deliveryStatus commonv1.DeliveryStatus
	// deliveryRequestedAt records when OMS successfully requested delivery.
	deliveryRequestedAt *time.Time
}

// NewOrderState creates a new OrderState instance with the given customer ID.
func NewOrderState(customerId uuid.UUID) *OrderState {
	return newOrderState(
		uuid.New(),
		customerId,
		make(Items, 0),
		OrderStatus_ORDER_STATUS_PENDING,
		0,
		nil,
		commonv1.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED,
		nil,
	)
}

// NewOrderStateFromPersisted builds an OrderState from persisted data (repository load).
// Single constructor for both "new order" and "reconstitute"; FSM rules live only here.
func NewOrderStateFromPersisted(
	id, customerId uuid.UUID,
	items Items,
	status OrderStatus,
	version int,
	deliveryInfo *DeliveryInfo,
	deliveryStatus commonv1.DeliveryStatus,
	deliveryRequestedAt *time.Time,
) *OrderState {
	if items == nil {
		items = make(Items, 0)
	}

	return newOrderState(id, customerId, items, status, version, deliveryInfo, deliveryStatus, deliveryRequestedAt)
}

// newOrderState is the single place that builds OrderState and configures the FSM.
func newOrderState(
	id, customerId uuid.UUID,
	items Items,
	status OrderStatus,
	version int,
	deliveryInfo *DeliveryInfo,
	deliveryStatus commonv1.DeliveryStatus,
	deliveryRequestedAt *time.Time,
) *OrderState {
	order := &OrderState{
		id:                  id,
		items:               items,
		customerId:          customerId,
		version:             version,
		domainEvents:        make([]domainevents.Event, 0),
		deliveryInfo:        deliveryInfo,
		deliveryStatus:      deliveryStatus,
		deliveryRequestedAt: cloneTimePointer(deliveryRequestedAt),
	}
	order.fsm = fsm.New(fsm.State(status.String()))
	order.addOrderTransitionRules(order.fsm)
	order.fsm.SetOnEnterState(order.onEnterState)
	order.fsm.SetOnExitState(order.onExitState)

	return order
}

// addOrderTransitionRules registers the order FSM transition rules (single source of truth).
// State = status (PENDING, PROCESSING, ...), Event = action (OrderTransitionEvent from proto).
//
//nolint:funcorder // unexported helper; order kept for FSM setup flow
func (o *OrderState) addOrderTransitionRules(f *fsm.FSM) {
	f.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PENDING.String()),
		fsm.Event(commonv1.OrderTransitionEvent_ORDER_TRANSITION_EVENT_CREATE.String()),
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
	)
	f.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PENDING.String()),
		fsm.Event(commonv1.OrderTransitionEvent_ORDER_TRANSITION_EVENT_CANCEL.String()),
		fsm.State(OrderStatus_ORDER_STATUS_CANCELED.String()),
	)
	f.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
		fsm.Event(commonv1.OrderTransitionEvent_ORDER_TRANSITION_EVENT_CANCEL.String()),
		fsm.State(OrderStatus_ORDER_STATUS_CANCELED.String()),
	)
	f.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
		fsm.Event(commonv1.OrderTransitionEvent_ORDER_TRANSITION_EVENT_COMPLETE.String()),
		fsm.State(OrderStatus_ORDER_STATUS_COMPLETED.String()),
	)
}

// GetVersion returns the current version for optimistic concurrency control.
func (o *OrderState) GetVersion() int {
	return o.version
}

// SetID sets the order ID (used when persisting a new order).
func (o *OrderState) SetID(id uuid.UUID) {
	o.id = id
}

// onEnterState is the callback executed when entering a new state.
// FSM is used only for transition validation; domain events are emitted in command methods (CreateOrder, CancelOrder, CompleteOrder).
func (o *OrderState) onEnterState(ctx context.Context, from, to fsm.State, event fsm.Event) { //nolint:funcorder // unexported FSM callback
	// No side effects here; events are raised by the aggregate in its command methods.
}

// onExitState is the callback executed when exiting a new state.
// Domain layer should not depend on infrastructure (logging, stdout, etc.).
func (o *OrderState) onExitState(ctx context.Context, from, to fsm.State, event fsm.Event) { //nolint:funcorder // unexported FSM callback
	// Domain layer should not perform side effects like logging.
}

// GetOrderID returns the unique identifier of the order.
func (o *OrderState) GetOrderID() uuid.UUID {
	return o.id
}

// GetItems returns a copy of the list of items in the order.
func (o *OrderState) GetItems() Items {
	o.mu.Lock()
	defer o.mu.Unlock()

	itemsCopy := make(Items, len(o.items))
	copy(itemsCopy, o.items)

	return itemsCopy
}

// GetCustomerId returns the customer ID associated with the order.
func (o *OrderState) GetCustomerId() uuid.UUID {
	return o.customerId
}

// GetDeliveryInfo returns the delivery information for the order.
func (o *OrderState) GetDeliveryInfo() *DeliveryInfo {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.deliveryInfo
}

// SetDeliveryInfo sets the delivery information for the order.
func (o *OrderState) SetDeliveryInfo(info DeliveryInfo) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !info.IsValid() {
		return ErrInvalidDeliveryInfo
	}

	currentStatus := o.getStatusUnlocked()
	if currentStatus == OrderStatus_ORDER_STATUS_COMPLETED ||
		currentStatus == OrderStatus_ORDER_STATUS_CANCELED {
		return &OrderTerminalStateError{Status: currentStatus}
	}

	if o.deliveryRequestedAt != nil || (o.deliveryInfo != nil && o.deliveryInfo.GetPackageId() != nil) {
		return &DeliveryAlreadyRequestedError{}
	}

	if o.deliveryStatus == commonv1.DeliveryStatus_DELIVERY_STATUS_ASSIGNED ||
		o.deliveryStatus == commonv1.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT ||
		o.deliveryStatus == commonv1.DeliveryStatus_DELIVERY_STATUS_DELIVERED ||
		o.deliveryStatus == commonv1.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED {
		return &DeliveryAlreadyInProgressError{DeliveryStatus: o.deliveryStatus}
	}

	o.deliveryInfo = &info

	return nil
}

// HasDeliveryInfo returns true if the order has delivery information.
func (o *OrderState) HasDeliveryInfo() bool {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.deliveryInfo != nil
}

// GetDeliveryStatus returns the current delivery status.
func (o *OrderState) GetDeliveryStatus() commonv1.DeliveryStatus {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.deliveryStatus
}

// GetDeliveryRequestedAt returns the time delivery was requested, if any.
func (o *OrderState) GetDeliveryRequestedAt() *time.Time {
	o.mu.Lock()
	defer o.mu.Unlock()

	return cloneTimePointer(o.deliveryRequestedAt)
}

// HasDeliveryRequest returns true once OMS successfully requested delivery.
func (o *OrderState) HasDeliveryRequest() bool {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.deliveryRequestedAt != nil
}

// SetDeliveryStatus updates the delivery status.
// Returns an error if the order is in a terminal state or if the transition is invalid.
func (o *OrderState) SetDeliveryStatus(status commonv1.DeliveryStatus) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.setDeliveryStatusLocked(status)
}

// RequestDelivery records a successful delivery request without changing delivery status.
func (o *OrderState) RequestDelivery(packageID *uuid.UUID, requestedAt time.Time) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if err := o.ensureDeliveryInfoLocked(); err != nil {
		return err
	}

	currentStatus := o.getStatusUnlocked()
	if currentStatus == OrderStatus_ORDER_STATUS_COMPLETED ||
		currentStatus == OrderStatus_ORDER_STATUS_CANCELED {
		return &OrderTerminalStateError{Status: currentStatus}
	}

	if o.deliveryRequestedAt != nil || o.deliveryInfo.GetPackageId() != nil {
		return &DeliveryAlreadyRequestedError{}
	}

	if requestedAt.IsZero() {
		requestedAt = time.Now()
	}

	if packageID != nil {
		o.deliveryInfo.SetPackageId(*packageID)
	}
	o.deliveryRequestedAt = cloneTimePointer(&requestedAt)

	ts := timestamppb.New(requestedAt)
	o.addDomainEvent(&eventsv1.OrderDeliveryRequestedEvent{
		OrderId:          o.id.String(),
		CustomerId:       o.customerId.String(),
		PickupAddress:    deliveryAddressToProto(o.deliveryInfo.GetPickupAddress()),
		DeliveryAddress:  deliveryAddressToProto(o.deliveryInfo.GetDeliveryAddress()),
		DeliveryPeriod:   deliveryPeriodToProto(o.deliveryInfo.GetDeliveryPeriod()),
		PackageInfo:      packageInfoToProto(o.deliveryInfo.GetPackageInfo()),
		Priority:         commonv1.DeliveryPriority(o.deliveryInfo.GetPriority()),
		CreatedAt:        ts,
		OccurredAt:       ts,
		PackageId:        packageIDString(packageID),
		AggregateVersion: o.nextAggregateVersion(),
	})

	return nil
}

// ApplyDeliveryAccepted updates the delivery lifecycle from Kafka truth.
func (o *OrderState) ApplyDeliveryAccepted(packageID *uuid.UUID, occurredAt time.Time) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if err := o.ensureDeliveryInfoLocked(); err != nil {
		return err
	}
	if err := o.syncPackageIDLocked(packageID); err != nil {
		return err
	}

	return o.applyDeliveryStatusUpdatedLocked(
		commonv1.DeliveryStatus_DELIVERY_STATUS_ACCEPTED,
		packageID,
		nil,
		occurredAt,
	)
}

// ApplyDeliveryAssigned updates the delivery lifecycle from Kafka truth.
func (o *OrderState) ApplyDeliveryAssigned(packageID *uuid.UUID, courierID *uuid.UUID, occurredAt time.Time) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if err := o.ensureDeliveryInfoLocked(); err != nil {
		return err
	}
	if err := o.syncPackageIDLocked(packageID); err != nil {
		return err
	}

	return o.applyDeliveryStatusUpdatedLocked(
		commonv1.DeliveryStatus_DELIVERY_STATUS_ASSIGNED,
		packageID,
		courierID,
		occurredAt,
	)
}

// ApplyDeliveryInTransit updates the delivery lifecycle from Kafka truth.
func (o *OrderState) ApplyDeliveryInTransit(packageID *uuid.UUID, courierID *uuid.UUID, occurredAt time.Time) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if err := o.ensureDeliveryInfoLocked(); err != nil {
		return err
	}
	if err := o.syncPackageIDLocked(packageID); err != nil {
		return err
	}

	return o.applyDeliveryStatusUpdatedLocked(
		commonv1.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT,
		packageID,
		courierID,
		occurredAt,
	)
}

// ApplyDeliveryDelivered marks delivery completed and closes the order.
func (o *OrderState) ApplyDeliveryDelivered(
	packageID *uuid.UUID,
	courierID *uuid.UUID,
	deliveryLocation *commonv1.DeliveryLocation,
	occurredAt time.Time,
) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if err := o.ensureDeliveryInfoLocked(); err != nil {
		return err
	}
	if err := o.syncPackageIDLocked(packageID); err != nil {
		return err
	}
	if err := o.setDeliveryStatusLocked(commonv1.DeliveryStatus_DELIVERY_STATUS_DELIVERED); err != nil {
		return err
	}

	ts := nonZeroEventTime(occurredAt)
	protoTS := timestamppb.New(ts)
	o.addDomainEvent(&eventsv1.OrderDeliveryCompletedEvent{
		OrderId:          o.id.String(),
		PackageId:        packageIDString(packageID),
		CourierId:        uuidString(courierID),
		DeliveredAt:      protoTS,
		DeliveryLocation: deliveryLocation,
		OccurredAt:       protoTS,
		AggregateVersion: o.nextAggregateVersion(),
	})

	return o.completeOrderLocked(ts)
}

// ApplyDeliveryFailed marks delivery failed and cancels the order.
func (o *OrderState) ApplyDeliveryFailed(
	packageID *uuid.UUID,
	courierID *uuid.UUID,
	details *commonv1.NotDeliveredDetails,
	occurredAt time.Time,
) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if err := o.ensureDeliveryInfoLocked(); err != nil {
		return err
	}
	if err := o.syncPackageIDLocked(packageID); err != nil {
		return err
	}
	if err := o.setDeliveryStatusLocked(commonv1.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED); err != nil {
		return err
	}

	ts := nonZeroEventTime(occurredAt)
	protoTS := timestamppb.New(ts)
	o.addDomainEvent(&eventsv1.OrderDeliveryFailedEvent{
		OrderId:             o.id.String(),
		PackageId:           packageIDString(packageID),
		CourierId:           uuidString(courierID),
		NotDeliveredDetails: details,
		FailedAt:            protoTS,
		OccurredAt:          protoTS,
		AggregateVersion:    o.nextAggregateVersion(),
	})

	return o.cancelOrderLocked("DELIVERY_FAILED", ts)
}

// isValidDeliveryStatusTransition checks if the delivery status transition is valid.
// Delivery status can only move forward: UNSPECIFIED -> ACCEPTED -> ASSIGNED -> IN_TRANSIT -> DELIVERED/NOT_DELIVERED
//
//nolint:funcorder // unexported helper
func (o *OrderState) isValidDeliveryStatusTransition(from, to commonv1.DeliveryStatus) bool {
	if from == commonv1.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED {
		return to == commonv1.DeliveryStatus_DELIVERY_STATUS_ACCEPTED
	}

	validTransitions := map[commonv1.DeliveryStatus][]commonv1.DeliveryStatus{
		commonv1.DeliveryStatus_DELIVERY_STATUS_ACCEPTED: {
			commonv1.DeliveryStatus_DELIVERY_STATUS_ASSIGNED,
		},
		commonv1.DeliveryStatus_DELIVERY_STATUS_ASSIGNED: {
			commonv1.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT,
		},
		commonv1.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT: {
			commonv1.DeliveryStatus_DELIVERY_STATUS_DELIVERED,
			commonv1.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED,
		},
	}

	allowedTargets, exists := validTransitions[from]
	if !exists {
		return false
	}

	return slices.Contains(allowedTargets, to)
}

// GetStatus returns the current status of the order.
func (o *OrderState) GetStatus() OrderStatus {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.getStatusUnlocked()
}

// getStatusUnlocked returns the current status without locking (for internal use).
//
//nolint:funcorder // unexported helper
func (o *OrderState) getStatusUnlocked() OrderStatus {
	return o.fsmStateToOrderStatus(o.fsm.GetCurrentState())
}

// fsmStateToOrderStatus converts FSM state to OrderStatus enum
//
//nolint:funcorder // unexported helper
func (o *OrderState) fsmStateToOrderStatus(state fsm.State) OrderStatus {
	for k, v := range OrderStatus_name {
		if v == state.String() {
			return OrderStatus(k)
		}
	}

	return OrderStatus_ORDER_STATUS_UNSPECIFIED
}

// orderItemsToProto converts domain Items to proto OrderItem slice for events.
func orderItemsToProto(items Items) []*commonv1.OrderItem {
	out := make([]*commonv1.OrderItem, 0, len(items))
	for _, it := range items {
		out = append(out, &commonv1.OrderItem{
			GoodId:   it.GetGoodId().String(),
			Quantity: it.GetQuantity(),
			Price:    it.GetPrice().String(),
		})
	}

	return out
}

// addDomainEvent adds a domain event (proto) to the aggregate's event list.
//
//nolint:funcorder // unexported helper
func (o *OrderState) addDomainEvent(event domainevents.Event) {
	o.domainEvents = append(o.domainEvents, event)
}

// GetDomainEvents returns all domain events that occurred during aggregate operations.
func (o *OrderState) GetDomainEvents() []domainevents.Event {
	o.mu.Lock()
	defer o.mu.Unlock()

	eventsCopy := make([]domainevents.Event, len(o.domainEvents))
	copy(eventsCopy, o.domainEvents)

	return eventsCopy
}

// ClearDomainEvents clears the domain events list.
func (o *OrderState) ClearDomainEvents() {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.domainEvents = o.domainEvents[:0]
}

// CreateOrder initializes the order with the provided items and transitions it to Processing state.
func (o *OrderState) CreateOrder(ctx context.Context, items Items) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if err := ValidateOrderItems(items); err != nil {
		return fmt.Errorf("cannot create order: %w", err)
	}

	itemsCopy := make(Items, len(items))
	copy(itemsCopy, items)

	currentStatus := o.getStatusUnlocked()
	if err := ValidateOrderStateTransition(currentStatus, OrderStatus_ORDER_STATUS_PROCESSING, itemsCopy); err != nil {
		return err
	}

	err := o.fsm.TriggerEvent(ctx, fsm.Event(commonv1.OrderTransitionEvent_ORDER_TRANSITION_EVENT_CREATE.String()))
	if err != nil {
		return err
	}

	o.items = itemsCopy

	ts := timestamppb.New(time.Now())
	o.addDomainEvent(&eventsv1.OrderCreated{
		OrderId:          o.id.String(),
		CustomerId:       o.customerId.String(),
		Items:            orderItemsToProto(o.items),
		Status:           OrderStatus_ORDER_STATUS_PROCESSING,
		CreatedAt:        ts,
		OccurredAt:       ts,
		AggregateVersion: o.nextAggregateVersion(),
	})

	return nil
}

// UpdateOrder updates the order's items.
func (o *OrderState) UpdateOrder(items Items) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	currentStatus := o.getStatusUnlocked()
	if currentStatus == OrderStatus_ORDER_STATUS_COMPLETED || currentStatus == OrderStatus_ORDER_STATUS_CANCELED {
		return &OrderTerminalStateError{Status: currentStatus}
	}

	canonical := make(map[uuid.UUID]Item, len(o.items)+len(items))
	for _, it := range o.items {
		canonical[it.GetGoodId()] = it
	}

	originalGoodIDs := make(map[uuid.UUID]bool, len(o.items))
	for _, it := range o.items {
		originalGoodIDs[it.GetGoodId()] = true
	}

	for _, item := range items {
		err := ValidateOrderItem(item)
		if err != nil {
			return fmt.Errorf("cannot update item %s: %w", item.GetGoodId(), err)
		}

		canonical[item.GetGoodId()] = item
	}

	result := make(Items, 0, len(canonical))
	for _, it := range o.items {
		result = append(result, canonical[it.GetGoodId()])
	}

	seenNew := make(map[uuid.UUID]bool)
	for _, it := range items {
		gid := it.GetGoodId()
		if !originalGoodIDs[gid] && !seenNew[gid] {
			seenNew[gid] = true
			result = append(result, canonical[gid])
		}
	}

	err := ValidateOrderItems(result)
	if err != nil {
		return fmt.Errorf("cannot update order: %w", err)
	}

	o.items = result

	return nil
}

// CancelOrder transitions the order to the Canceled state.
func (o *OrderState) CancelOrder() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.cancelOrderLocked("", time.Now())
}

// CompleteOrder transitions the order to the Completed state.
func (o *OrderState) CompleteOrder() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.completeOrderLocked(time.Now())
}

func (o *OrderState) setDeliveryStatusLocked(status commonv1.DeliveryStatus) error {
	currentOrderStatus := o.getStatusUnlocked()
	if currentOrderStatus == OrderStatus_ORDER_STATUS_COMPLETED ||
		currentOrderStatus == OrderStatus_ORDER_STATUS_CANCELED {
		return &OrderTerminalStateError{Status: currentOrderStatus}
	}

	if !o.isValidDeliveryStatusTransition(o.deliveryStatus, status) {
		return &InvalidDeliveryStatusTransitionError{From: o.deliveryStatus, To: status}
	}

	o.deliveryStatus = status

	return nil
}

func (o *OrderState) applyDeliveryStatusUpdatedLocked(
	status commonv1.DeliveryStatus,
	packageID *uuid.UUID,
	courierID *uuid.UUID,
	occurredAt time.Time,
) error {
	if err := o.setDeliveryStatusLocked(status); err != nil {
		return err
	}

	ts := nonZeroEventTime(occurredAt)
	protoTS := timestamppb.New(ts)
	o.addDomainEvent(&eventsv1.OrderDeliveryStatusUpdatedEvent{
		OrderId:          o.id.String(),
		PackageId:        packageIDString(packageID),
		Status:           status,
		UpdatedAt:        protoTS,
		CourierId:        uuidString(courierID),
		OccurredAt:       protoTS,
		AggregateVersion: o.nextAggregateVersion(),
	})

	return nil
}

func (o *OrderState) cancelOrderLocked(reason string, occurredAt time.Time) error {
	err := o.fsm.TriggerEvent(context.Background(), fsm.Event(commonv1.OrderTransitionEvent_ORDER_TRANSITION_EVENT_CANCEL.String()))
	if err != nil {
		return err
	}

	ts := timestamppb.New(nonZeroEventTime(occurredAt))
	o.addDomainEvent(&eventsv1.OrderCancelled{
		OrderId:          o.id.String(),
		CustomerId:       o.customerId.String(),
		Status:           OrderStatus_ORDER_STATUS_CANCELED,
		Reason:           reason,
		CancelledAt:      ts,
		OccurredAt:       ts,
		AggregateVersion: o.nextAggregateVersion(),
	})

	return nil
}

func (o *OrderState) completeOrderLocked(occurredAt time.Time) error {
	currentStatus := o.getStatusUnlocked()
	if currentStatus != OrderStatus_ORDER_STATUS_PROCESSING {
		return &InvalidOrderTransitionError{From: currentStatus, To: OrderStatus_ORDER_STATUS_COMPLETED}
	}

	if err := ValidateOrderItems(o.items); err != nil {
		return fmt.Errorf("cannot complete order with invalid items: %w", err)
	}

	err := o.fsm.TriggerEvent(context.Background(), fsm.Event(commonv1.OrderTransitionEvent_ORDER_TRANSITION_EVENT_COMPLETE.String()))
	if err != nil {
		return err
	}

	ts := timestamppb.New(nonZeroEventTime(occurredAt))
	o.addDomainEvent(&eventsv1.OrderCompleted{
		OrderId:          o.id.String(),
		CustomerId:       o.customerId.String(),
		Status:           OrderStatus_ORDER_STATUS_COMPLETED,
		CompletedAt:      ts,
		OccurredAt:       ts,
		AggregateVersion: o.nextAggregateVersion(),
	})

	return nil
}

func (o *OrderState) ensureDeliveryInfoLocked() error {
	if o.deliveryInfo == nil {
		return ErrDeliveryInfoRequired
	}

	return nil
}

func (o *OrderState) syncPackageIDLocked(packageID *uuid.UUID) error {
	if packageID == nil {
		return nil
	}
	if err := o.ensureDeliveryInfoLocked(); err != nil {
		return err
	}

	currentPackageID := o.deliveryInfo.GetPackageId()
	if currentPackageID == nil {
		o.deliveryInfo.SetPackageId(*packageID)
		return nil
	}

	if *currentPackageID != *packageID {
		return &DeliveryPackageMismatchError{
			Expected: currentPackageID.String(),
			Actual:   packageID.String(),
		}
	}

	return nil
}

func (o *OrderState) nextAggregateVersion() int32 {
	return int32(o.version + 1)
}

func cloneTimePointer(in *time.Time) *time.Time {
	if in == nil {
		return nil
	}

	cloned := *in
	return &cloned
}

func nonZeroEventTime(ts time.Time) time.Time {
	if ts.IsZero() {
		return time.Now()
	}

	return ts
}

func packageIDString(packageID *uuid.UUID) string {
	if packageID == nil {
		return ""
	}

	return packageID.String()
}

func uuidString(value *uuid.UUID) string {
	if value == nil {
		return ""
	}

	return value.String()
}

func deliveryAddressToProto(addr addressvo.Address) *commonv1.DeliveryAddress {
	return &commonv1.DeliveryAddress{
		Street:     addr.Street(),
		City:       addr.City(),
		PostalCode: addr.PostalCode(),
		Country:    addr.Country(),
		Latitude:   addr.Latitude(),
		Longitude:  addr.Longitude(),
	}
}

func deliveryPeriodToProto(period DeliveryPeriod) *commonv1.DeliveryPeriod {
	return &commonv1.DeliveryPeriod{
		StartTime: timestamppb.New(period.GetStartTime()),
		EndTime:   timestamppb.New(period.GetEndTime()),
	}
}

func packageInfoToProto(info PackageInfo) *commonv1.PackageInfo {
	return &commonv1.PackageInfo{
		WeightKg: info.GetWeightKg(),
	}
}
