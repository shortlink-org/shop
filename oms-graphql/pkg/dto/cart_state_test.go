package dto //nolint:testpackage // testing exported API only

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cartmodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/cart/v1/model/v1"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

func TestCartStateToService(t *testing.T) {
	t.Parallel()

	t.Run("nil returns nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, CartStateToService(nil))
	})

	t.Run("maps cart state and items", func(t *testing.T) {
		t.Parallel()

		in := &cartmodel.CartState{
			CartId: "cart-1",
			Items: []*cartmodel.CartItem{
				{GoodId: "good-a", Quantity: 2},
				{GoodId: "good-b", Quantity: 1},
			},
		}
		out := CartStateToService(in)
		assert.NotNil(t, out)
		assert.Equal(t, "cart-1", out.GetCartId().GetValue())
		requireList(t, out.GetItems())
		items := out.GetItems().GetList().GetItems()
		assert.Len(t, items, 2)
		assert.Equal(t, "good-a", items[0].GetGoodId().GetValue())
		assert.Equal(t, int32(2), items[0].GetQuantity().GetValue())
		assert.Equal(t, "good-b", items[1].GetGoodId().GetValue())
		assert.Equal(t, int32(1), items[1].GetQuantity().GetValue())
	})
}

func requireList(t *testing.T, items *servicepb.ListOfCartItem) {
	t.Helper()

	if items == nil || items.GetList() == nil {
		t.Fatal("expected non-nil items list")
	}
}
