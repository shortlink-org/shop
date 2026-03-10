package dto

import (
	"time"

	"google.golang.org/protobuf/types/known/wrapperspb"

	commonpb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/domain/order/v1/common"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

// DeliveryPeriodToService maps OMS delivery period to Connect response.
func DeliveryPeriodToService(period *commonpb.DeliveryPeriod) *servicepb.DeliveryPeriod {
	if period == nil {
		return nil
	}

	return &servicepb.DeliveryPeriod{
		StartTime: wrapperspb.String(period.GetStartTime().AsTime().Format(time.RFC3339)),
		EndTime:   wrapperspb.String(period.GetEndTime().AsTime().Format(time.RFC3339)),
	}
}

// ParseTimestamp parses RFC3339 timestamp string. Returns InvalidArgument error on failure.
func ParseTimestamp(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return parsed, nil
	}

	return time.Time{}, InvalidArgument("deliveryPeriod values must be RFC3339 timestamps")
}
