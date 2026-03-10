package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

func TestCartItemInputsToOMS(t *testing.T) {
	t.Parallel()

	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, CartItemInputsToOMS(nil))
	})

	t.Run("maps items", func(t *testing.T) {
		in := &servicepb.ItemRequest{
			Items: []*servicepb.CartItemInput{
				{GoodId: "g1", Quantity: 3},
				{GoodId: "g2", Quantity: 1},
			},
		}
		out := CartItemInputsToOMS(in)
		assert.Len(t, out, 2)
		assert.Equal(t, "g1", out[0].GetGoodId())
		assert.Equal(t, int32(3), out[0].GetQuantity())
		assert.Equal(t, "g2", out[1].GetGoodId())
		assert.Equal(t, int32(1), out[1].GetQuantity())
	})
}
