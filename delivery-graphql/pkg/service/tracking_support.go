package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/99designs/gqlgen/graphql"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	deliverygrpc "github.com/shortlink-org/shop/delivery-graphql/pkg/generated/infrastructure/rpc/delivery/v1"
)

const (
	authHeader        = "Authorization"
	xUserIDHeader     = "X-User-ID"
	traceparentHeader = "traceparent"
	traceIDHeader     = "trace-id"
)

var errMissingTrackingIdentity = errors.New("missing required Authorization or X-User-ID header")

func trackingContextFromHeaders(ctx context.Context, headers http.Header) (context.Context, error) {
	if strings.TrimSpace(headers.Get(authHeader)) == "" && strings.TrimSpace(headers.Get(xUserIDHeader)) == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errMissingTrackingIdentity)
	}

	return forwardContextFromHeaders(ctx, headers), nil
}

func forwardContextFromHeaders(ctx context.Context, headers http.Header) context.Context {
	mdPairs := make([]string, 0, 8)

	for _, key := range []string{authHeader, xUserIDHeader, traceparentHeader, traceIDHeader} {
		if value := strings.TrimSpace(headers.Get(key)); value != "" {
			mdPairs = append(mdPairs, strings.ToLower(key), value)
		}
	}

	if len(mdPairs) == 0 {
		return ctx
	}

	return metadata.NewOutgoingContext(ctx, metadata.Pairs(mdPairs...))
}

func operationHeaders(ctx context.Context) http.Header {
	oc := graphql.GetOperationContext(ctx)
	if oc == nil || oc.Headers == nil {
		return http.Header{}
	}

	return oc.Headers
}

func stringPtr(value string) *string {
	return &value
}

func int32PtrToInt(value *int32) *int {
	if value == nil {
		return nil
	}

	v := int(*value)
	return &v
}

func timestampString(value *timestamppb.Timestamp) *string {
	if value == nil {
		return nil
	}

	return stringPtr(value.AsTime().UTC().Format(time.RFC3339))
}

func randomAddressToGraph(resp *deliverygrpc.GetRandomAddressResponse) *RandomAddressResponse {
	if resp == nil || resp.GetAddress() == nil {
		return &RandomAddressResponse{}
	}

	addr := resp.GetAddress()
	return &RandomAddressResponse{
		Address: &Address{
			Street:     stringPtr(addr.GetStreet()),
			City:       stringPtr(addr.GetCity()),
			PostalCode: stringPtr(addr.GetPostalCode()),
			Country:    stringPtr(addr.GetCountry()),
			Latitude:   &addr.Latitude,
			Longitude:  &addr.Longitude,
		},
	}
}

func deliveryTrackingToGraph(resp *deliverygrpc.GetOrderTrackingResponse) *DeliveryTracking {
	if resp == nil {
		return nil
	}

	var courier *DeliveryCourier
	if source := resp.GetCourier(); source != nil {
		courier = &DeliveryCourier{
			CourierID:     stringPtr(source.GetCourierId()),
			Name:          stringPtr(source.GetName()),
			Phone:         stringPtr(source.GetPhone()),
			TransportType: stringPtr(enumName(source.GetTransportType().String(), "TRANSPORT_TYPE_")),
			Status:        stringPtr(enumName(source.GetStatus().String(), "COURIER_STATUS_")),
			LastActiveAt:  timestampString(source.GetLastActiveAt()),
		}
		if location := source.GetCurrentLocation(); location != nil {
			courier.CurrentLocation = &DeliveryLocation{
				Latitude:  &location.Latitude,
				Longitude: &location.Longitude,
			}
		}
	}

	return &DeliveryTracking{
		OrderID:                   stringPtr(resp.GetOrderId()),
		PackageID:                 stringPtr(resp.GetPackageId()),
		Status:                    stringPtr(enumName(resp.GetStatus().String(), "PACKAGE_STATUS_")),
		Courier:                   courier,
		EstimatedMinutesRemaining: int32PtrToInt(resp.EstimatedMinutesRemaining),
		DistanceKmRemaining:       resp.DistanceKmRemaining,
		EstimatedArrivalAt:        timestampString(resp.GetEstimatedArrivalAt()),
		AssignedAt:                timestampString(resp.GetAssignedAt()),
		DeliveredAt:               timestampString(resp.GetDeliveredAt()),
	}
}

func trackingFingerprint(tracking *DeliveryTracking) string {
	if tracking == nil {
		return ""
	}

	payload, err := json.Marshal(tracking)
	if err != nil {
		return ""
	}

	return string(payload)
}
