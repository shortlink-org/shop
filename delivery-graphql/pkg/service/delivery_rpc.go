package service

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	deliverygrpc "github.com/shortlink-org/shop/delivery-graphql/pkg/generated/infrastructure/rpc/delivery/v1"
)

func (s *Service) getRandomAddress(ctx context.Context, headers http.Header) (*deliverygrpc.GetRandomAddressResponse, error) {
	return s.deliveryClient.GetRandomAddress(forwardContextFromHeaders(ctx, headers), &deliverygrpc.GetRandomAddressRequest{})
}

func (s *Service) getOrderTracking(
	ctx context.Context,
	headers http.Header,
	orderID string,
) (*deliverygrpc.GetOrderTrackingResponse, error) {
	outboundCtx, err := trackingContextFromHeaders(ctx, headers)
	if err != nil {
		return nil, err
	}

	resp, err := s.deliveryClient.GetOrderTracking(outboundCtx, &deliverygrpc.GetOrderTrackingRequest{
		OrderId: orderID,
	})
	if err != nil {
		st, ok := grpcstatus.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	return resp, nil
}

func (s *Service) subscribeOrderTracking(
	ctx context.Context,
	headers http.Header,
	orderID string,
) (grpc.ServerStreamingClient[deliverygrpc.GetOrderTrackingResponse], error) {
	outboundCtx, err := trackingContextFromHeaders(ctx, headers)
	if err != nil {
		return nil, err
	}

	return s.deliveryClient.SubscribeOrderTracking(outboundCtx, &deliverygrpc.GetOrderTrackingRequest{
		OrderId: orderID,
	})
}
