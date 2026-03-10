package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cartmodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/cart/v1/model/v1"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

func TestCartStateToService(t *testing.T) {
	t.Parallel()

	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, CartStateToService(nil))
	})

	t.Run("maps cart state and items", func(t *testing.T) {
		in := &cartmodel.CartState{
			CartId: "cart-1",
			Items: []*cartmodel.CartItem{
				{GoodId: "good-a", Quantity: 2},
				{GoodId: "good-b", Quantity: 1},
			},
		}
		out := CartStateToService(in)
		assert.NotNil(t, out)
		assert.Equal(t, "cart-1", out.CartId.GetValue())
		requireList(t, out.Items)
		items := out.Items.GetList().GetItems()
		assert.Len(t, items, 2)
		assert.Equal(t, "good-a", items[0].GoodId.GetValue())
		assert.Equal(t, int32(2), items[0].Quantity.GetValue())
		assert.Equal(t, "good-b", items[1].GoodId.GetValue())
		assert.Equal(t, int32(1), items[1].Quantity.GetValue())
	})
}

func requireList(t *testing.T, items *servicepb.ListOfCartItem) {
	t.Helper()
	if items == nil || items.GetList() == nil {
		t.Fatal("expected non-nil items list")
	}
}
