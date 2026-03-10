package dto //nolint:testpackage // testing exported API only

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonpb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/domain/order/v1/common"
	ordermodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/order/v1/model/v1"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

func TestOrderStateToService(t *testing.T) {
	t.Parallel()

	t.Run("nil returns nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, OrderStateToService(nil))
	})

	t.Run("maps order state", func(t *testing.T) {
		t.Parallel()

		requestedAt := time.Date(2026, time.March, 11, 12, 30, 0, 0, time.UTC)
		in := &ordermodel.OrderState{
			Id:             "order-1",
			Status:         commonpb.OrderStatus_ORDER_STATUS_PENDING,
			DeliveryStatus: commonpb.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT,
			PackageId:      "package-1",
			RequestedAt:    timestamppb.New(requestedAt),
			Items: []*ordermodel.OrderItem{
				{Id: "item-1", Quantity: 2, Price: 10.5},
			},
			DeliveryInfo: &commonpb.DeliveryInfo{
				PackageInfo: &commonpb.PackageInfo{WeightKg: 1.0},
			},
		}
		out := OrderStateToService(in)
		assert.NotNil(t, out)
		assert.Equal(t, "order-1", out.GetId().GetValue())
		assert.Equal(t, "ORDER_STATUS_PENDING", out.GetStatus().GetValue())
		assert.Equal(t, "DELIVERY_STATUS_IN_TRANSIT", out.GetDeliveryStatus().GetValue())
		assert.Equal(t, "package-1", out.GetPackageId().GetValue())
		assert.True(t, out.GetRequestedAt().AsTime().Equal(requestedAt))
		requireOrderItems(t, out.GetItems())
		items := out.GetItems().GetList().GetItems()
		assert.Len(t, items, 1)
		assert.InEpsilon(t, 10.5, items[0].GetPrice().GetValue(), 1e-9)
		assert.NotNil(t, out.GetDeliveryInfo())
	})
}

func requireOrderItems(t *testing.T, items *servicepb.ListOfOrderItem) {
	t.Helper()

	if items == nil || items.GetList() == nil {
		t.Fatal("expected non-nil order items list")
	}
}
