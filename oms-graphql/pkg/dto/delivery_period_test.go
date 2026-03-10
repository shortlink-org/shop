package dto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonpb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/domain/order/v1/common"
)

func TestDeliveryPeriodToService(t *testing.T) {
	t.Parallel()
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, DeliveryPeriodToService(nil))
	})
	t.Run("maps period", func(t *testing.T) {
		ts := timestamppb.New(time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC))
		in := &commonpb.DeliveryPeriod{StartTime: ts, EndTime: ts}
		out := DeliveryPeriodToService(in)
		assert.Contains(t, out.StartTime.GetValue(), "2025-01-15")
	})
}

func TestParseTimestamp(t *testing.T) {
	t.Parallel()
	t.Run("valid RFC3339", func(t *testing.T) {
		got, err := ParseTimestamp("2025-01-15T10:00:00Z")
		assert.NoError(t, err)
		assert.True(t, got.Year() == 2025 && got.Month() == 1 && got.Day() == 15)
	})
	t.Run("invalid returns error", func(t *testing.T) {
		_, err := ParseTimestamp("not-a-date")
		assert.Error(t, err)
	})
}
