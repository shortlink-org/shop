package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

func TestCreateOrderRequestFromInput(t *testing.T) {
	t.Parallel()

	t.Run("nil returns error", func(t *testing.T) {
		_, err := CreateOrderRequestFromInput(nil)
		assert.Error(t, err)
	})

	t.Run("maps input and generates order id", func(t *testing.T) {
		in := &servicepb.CreateOrderInput{
			Items: []*servicepb.OrderItemInput{
				{Id: "good-1", Quantity: 2, Price: 99.5},
			},
		}
		out, err := CreateOrderRequestFromInput(in)
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.NotEmpty(t, out.GetOrder().GetId())
		assert.Len(t, out.GetOrder().GetItems(), 1)
		assert.Equal(t, "good-1", out.GetOrder().GetItems()[0].GetId())
		assert.Equal(t, int32(2), out.GetOrder().GetItems()[0].GetQuantity())
		assert.Equal(t, 99.5, out.GetOrder().GetItems()[0].GetPrice())
	})
}

func TestCreateOrderRequestFromInput_EmptyItems(t *testing.T) {
	t.Parallel()
	in := &servicepb.CreateOrderInput{Items: []*servicepb.OrderItemInput{}}
	out, err := CreateOrderRequestFromInput(in)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.NotEmpty(t, out.GetOrder().GetId())
	assert.Empty(t, out.GetOrder().GetItems())
}
