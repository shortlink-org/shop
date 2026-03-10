package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	commonpb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/domain/order/v1/common"
	ordermodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/order/v1/model/v1"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

func TestOrderStateToService(t *testing.T) {
	t.Parallel()

	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, OrderStateToService(nil))
	})

	t.Run("maps order state", func(t *testing.T) {
		in := &ordermodel.OrderState{
			Id:     "order-1",
			Status: commonpb.OrderStatus_ORDER_STATUS_PENDING,
			Items: []*ordermodel.OrderItem{
				{Id: "item-1", Quantity: 2, Price: 10.5},
			},
			DeliveryInfo: &commonpb.DeliveryInfo{
				PackageInfo: &commonpb.PackageInfo{WeightKg: 1.0},
			},
		}
		out := OrderStateToService(in)
		assert.NotNil(t, out)
		assert.Equal(t, "order-1", out.Id.GetValue())
		assert.Equal(t, "ORDER_STATUS_PENDING", out.Status.GetValue())
		requireOrderItems(t, out.Items)
		items := out.Items.GetList().GetItems()
		assert.Len(t, items, 1)
		assert.Equal(t, 10.5, items[0].Price.GetValue())
		assert.NotNil(t, out.DeliveryInfo)
	})
}

func requireOrderItems(t *testing.T, items *servicepb.ListOfOrderItem) {
	t.Helper()
	if items == nil || items.GetList() == nil {
		t.Fatal("expected non-nil order items list")
	}
}
