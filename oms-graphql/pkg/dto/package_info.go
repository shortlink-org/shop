package dto

import (
	commonpb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/domain/order/v1/common"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// PackageInfoToService maps OMS package info to Connect response.
func PackageInfoToService(info *commonpb.PackageInfo) *servicepb.PackageInfo {
	if info == nil {
		return nil
	}
	return &servicepb.PackageInfo{
		WeightKg: wrapperspb.Double(info.GetWeightKg()),
	}
}
