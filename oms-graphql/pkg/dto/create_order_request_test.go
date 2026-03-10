package dto //nolint:testpackage // testing exported API only

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

func TestCreateOrderRequestFromInput(t *testing.T) {
	t.Parallel()

	t.Run("nil returns error", func(t *testing.T) {
		t.Parallel()

		_, err := CreateOrderRequestFromInput(nil)
		assert.Error(t, err)
	})

	t.Run("maps input and generates order id", func(t *testing.T) {
		t.Parallel()

		in := &servicepb.CreateOrderInput{
			Items: []*servicepb.OrderItemInput{
				{Id: "good-1", Quantity: 2, Price: 99.5},
			},
		}
		out, err := CreateOrderRequestFromInput(in)
		require.NoError(t, err)
		assert.NotNil(t, out)
		assert.NotEmpty(t, out.GetOrder().GetId())
		assert.Len(t, out.GetOrder().GetItems(), 1)
		assert.Equal(t, "good-1", out.GetOrder().GetItems()[0].GetId())
		assert.Equal(t, int32(2), out.GetOrder().GetItems()[0].GetQuantity())
		assert.InEpsilon(t, 99.5, out.GetOrder().GetItems()[0].GetPrice(), 1e-9)
	})
}

func TestCreateOrderRequestFromInput_EmptyItems(t *testing.T) {
	t.Parallel()

	in := &servicepb.CreateOrderInput{Items: []*servicepb.OrderItemInput{}}
	out, err := CreateOrderRequestFromInput(in)
	require.NoError(t, err)
	assert.NotNil(t, out)
	assert.NotEmpty(t, out.GetOrder().GetId())
	assert.Empty(t, out.GetOrder().GetItems())
}
