package get

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

type stubUnitOfWork struct{}

func (stubUnitOfWork) Begin(ctx context.Context) (context.Context, error) { return ctx, nil }
func (stubUnitOfWork) Commit(context.Context) error                       { return nil }
func (stubUnitOfWork) Rollback(context.Context) error                     { return nil }

type stubOrderRepository struct {
	order *orderv1.OrderState
	err   error
}

func (s stubOrderRepository) Load(context.Context, uuid.UUID) (*orderv1.OrderState, error) {
	return s.order, s.err
}

func (stubOrderRepository) LoadByPackageID(context.Context, uuid.UUID) (*orderv1.OrderState, error) {
	panic("unexpected call")
}

func (stubOrderRepository) Save(context.Context, *orderv1.OrderState) error {
	panic("unexpected call")
}

func (stubOrderRepository) List(context.Context, ports.ListFilter) ([]*orderv1.OrderState, error) {
	panic("unexpected call")
}

func (stubOrderRepository) ListByCustomer(context.Context, uuid.UUID) ([]*orderv1.OrderState, error) {
	panic("unexpected call")
}

func TestHandleReturnsNotFoundForDifferentCustomer(t *testing.T) {
	t.Parallel()

	orderID := uuid.New()
	ownerID := uuid.New()
	requestorID := uuid.New()
	handler, err := NewHandler(
		stubUnitOfWork{},
		stubOrderRepository{
			order: orderv1.NewOrderStateFromPersisted(
				orderID,
				ownerID,
				nil,
				orderv1.OrderStatus_ORDER_STATUS_PENDING,
				1,
				nil,
				commonv1.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED,
				nil,
			),
		},
	)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	_, err = handler.Handle(context.Background(), NewCustomerScopedQuery(orderID, requestorID))
	if !errors.Is(err, ports.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestHandleReturnsOrderForOwner(t *testing.T) {
	t.Parallel()

	orderID := uuid.New()
	ownerID := uuid.New()
	expected := orderv1.NewOrderStateFromPersisted(
		orderID,
		ownerID,
		nil,
		orderv1.OrderStatus_ORDER_STATUS_PENDING,
		1,
		nil,
		commonv1.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED,
		nil,
	)

	handler, err := NewHandler(stubUnitOfWork{}, stubOrderRepository{order: expected})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	result, err := handler.Handle(context.Background(), NewCustomerScopedQuery(orderID, ownerID))
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if result.GetOrderID() != expected.GetOrderID() {
		t.Fatalf("expected order %s, got %s", expected.GetOrderID(), result.GetOrderID())
	}
}
