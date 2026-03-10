package dto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	commonpb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/domain/order/v1/common"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

func TestDeliveryInfoToService(t *testing.T) {
	t.Parallel()

	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, DeliveryInfoToService(nil))
	})

	t.Run("maps delivery info", func(t *testing.T) {
		ts := timestamppb.New(time.Now())
		in := &commonpb.DeliveryInfo{
			PickupAddress: &commonpb.DeliveryAddress{Street: "s1", City: "c1", Country: "RU"},
			DeliveryPeriod: &commonpb.DeliveryPeriod{
				StartTime: ts,
				EndTime:   ts,
			},
			PackageInfo: &commonpb.PackageInfo{WeightKg: 2.5},
			Priority:    commonpb.DeliveryPriority_DELIVERY_PRIORITY_NORMAL,
		}
		out := DeliveryInfoToService(in)
		assert.NotNil(t, out)
		assert.NotNil(t, out.PickupAddress)
		assert.Equal(t, "s1", out.PickupAddress.Street.GetValue())
		assert.Equal(t, 2.5, out.PackageInfo.WeightKg.GetValue())
	})
}

func TestDeliveryInfoFromInput(t *testing.T) {
	t.Parallel()

	t.Run("nil returns nil", func(t *testing.T) {
		out, err := DeliveryInfoFromInput(nil)
		assert.NoError(t, err)
		assert.Nil(t, out)
	})

	t.Run("valid RFC3339 timestamps", func(t *testing.T) {
		in := &servicepb.DeliveryInfoInput{
			DeliveryPeriod: &servicepb.DeliveryPeriodInput{
				StartTime: "2025-01-15T10:00:00Z",
				EndTime:   "2025-01-15T18:00:00Z",
			},
			PackageInfo: &servicepb.PackageInfoInput{WeightKg: 1.0},
			Priority:    wrapperspb.String("NORMAL"),
		}
		out, err := DeliveryInfoFromInput(in)
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, 1.0, out.PackageInfo.GetWeightKg())
	})

	t.Run("invalid timestamp returns error", func(t *testing.T) {
		in := &servicepb.DeliveryInfoInput{
			DeliveryPeriod: &servicepb.DeliveryPeriodInput{
				StartTime: "not-a-date",
				EndTime:   "2025-01-15T18:00:00Z",
			},
			PackageInfo: &servicepb.PackageInfoInput{WeightKg: 1.0},
		}
		_, err := DeliveryInfoFromInput(in)
		assert.Error(t, err)
	})
}
