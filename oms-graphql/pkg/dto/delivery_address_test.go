package dto //nolint:testpackage // testing exported API only

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"

	commonpb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/domain/order/v1/common"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

func TestDeliveryAddressToService(t *testing.T) {
	t.Parallel()
	t.Run("nil returns nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, DeliveryAddressToService(nil))
	})
	t.Run("maps address", func(t *testing.T) {
		t.Parallel()

		in := &commonpb.DeliveryAddress{
			Street: "s1", City: "c1", PostalCode: "123", Country: "RU",
			Latitude: 1.0, Longitude: 2.0,
		}
		out := DeliveryAddressToService(in)
		assert.Equal(t, "s1", out.GetStreet().GetValue())
		assert.Equal(t, "123", out.GetPostalCode().GetValue())
		assert.InEpsilon(t, 1.0, out.GetLatitude().GetValue(), 1e-9)
	})
}

func TestDeliveryAddressFromInput(t *testing.T) {
	t.Parallel()
	t.Run("nil returns nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, DeliveryAddressFromInput(nil))
	})
	t.Run("maps input", func(t *testing.T) {
		t.Parallel()

		in := &servicepb.DeliveryAddressInput{
			Street: "s1", City: "c1", Country: "RU",
			PostalCode: wrapperspb.String("456"),
			Latitude:   wrapperspb.Double(3.0),
		}
		out := DeliveryAddressFromInput(in)
		assert.Equal(t, "s1", out.GetStreet())
		assert.Equal(t, "456", out.GetPostalCode())
		assert.InEpsilon(t, 3.0, out.GetLatitude(), 1e-9)
	})
}
