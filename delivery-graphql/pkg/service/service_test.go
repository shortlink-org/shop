package service //nolint:testpackage // testing exported API only

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	grpcstatus "google.golang.org/grpc/status"

	deliverygrpc "github.com/shortlink-org/shop/delivery-graphql/pkg/generated/infrastructure/rpc/delivery/v1"
	servicepb "github.com/shortlink-org/shop/delivery-graphql/pkg/generated/service/v1"
	serviceconnect "github.com/shortlink-org/shop/delivery-graphql/pkg/generated/service/v1/v1connect"
)

func TestQueryRandomAddressMapsAddress(t *testing.T) {
	t.Parallel()

	client := &stubDeliveryClient{
		getRandomAddressFunc: func(ctx context.Context, _ *deliverygrpc.GetRandomAddressRequest, _ ...grpc.CallOption) (*deliverygrpc.GetRandomAddressResponse, error) {
			return &deliverygrpc.GetRandomAddressResponse{
				Address: &deliverygrpc.Address{
					Street:    "Alexanderplatz 1",
					City:      "Berlin",
					Country:   "Germany",
					Latitude:  52.5219,
					Longitude: 13.4132,
				},
			}, nil
		},
	}

	svc := New(nil, client)
	req := connect.NewRequest(&servicepb.QueryRandomAddressRequest{})

	resp, err := svc.QueryRandomAddress(context.Background(), req)
	if err != nil {
		t.Fatalf("QueryRandomAddress returned error: %v", err)
	}

	addr := resp.Msg.GetRandomAddress().GetAddress()
	if addr == nil {
		t.Fatal("expected address in response")
	}

	if addr.GetStreet() != "Alexanderplatz 1" {
		t.Fatalf("expected street Alexanderplatz 1, got %s", addr.GetStreet())
	}
	if addr.GetCity() != "Berlin" {
		t.Fatalf("expected city Berlin, got %s", addr.GetCity())
	}
	if addr.GetCountry() != "Germany" {
		t.Fatalf("expected country Germany, got %s", addr.GetCountry())
	}
	if addr.GetLatitude() != 52.5219 {
		t.Fatalf("expected latitude 52.5219, got %v", addr.GetLatitude())
	}
	if addr.GetLongitude() != 13.4132 {
		t.Fatalf("expected longitude 13.4132, got %v", addr.GetLongitude())
	}
}

func TestQueryRandomAddressNilAddressReturnsEmpty(t *testing.T) {
	t.Parallel()

	client := &stubDeliveryClient{
		getRandomAddressFunc: func(ctx context.Context, _ *deliverygrpc.GetRandomAddressRequest, _ ...grpc.CallOption) (*deliverygrpc.GetRandomAddressResponse, error) {
			return &deliverygrpc.GetRandomAddressResponse{}, nil
		},
	}

	svc := New(nil, client)
	req := connect.NewRequest(&servicepb.QueryRandomAddressRequest{})

	resp, err := svc.QueryRandomAddress(context.Background(), req)
	if err != nil {
		t.Fatalf("QueryRandomAddress returned error: %v", err)
	}

	ra := resp.Msg.GetRandomAddress()
	if ra == nil {
		t.Fatal("expected RandomAddressResponse")
	}
	if ra.GetAddress() != nil && (ra.GetAddress().GetStreet() != "" || ra.GetAddress().GetCity() != "") {
		t.Fatalf("expected empty address, got %+v", ra.GetAddress())
	}
}

func TestQueryRandomAddressPropagatesGRPCError(t *testing.T) {
	t.Parallel()

	client := &stubDeliveryClient{
		getRandomAddressFunc: func(ctx context.Context, _ *deliverygrpc.GetRandomAddressRequest, _ ...grpc.CallOption) (*deliverygrpc.GetRandomAddressResponse, error) {
			return nil, grpcstatus.Error(grpccodes.Unavailable, "delivery unavailable")
		},
	}

	svc := New(nil, client)
	req := connect.NewRequest(&servicepb.QueryRandomAddressRequest{})

	_, err := svc.QueryRandomAddress(context.Background(), req)
	if err == nil {
		t.Fatal("expected error")
	}

	if code := connect.CodeOf(err); code != connect.CodeUnavailable {
		t.Fatalf("expected CodeUnavailable, got %v", code)
	}
}

func TestDeliveryHandlerQueryRandomAddressOverConnect(t *testing.T) {
	t.Parallel()

	client := &stubDeliveryClient{
		getRandomAddressFunc: func(ctx context.Context, _ *deliverygrpc.GetRandomAddressRequest, _ ...grpc.CallOption) (*deliverygrpc.GetRandomAddressResponse, error) {
			return &deliverygrpc.GetRandomAddressResponse{
				Address: &deliverygrpc.Address{
					Street:  "Kurfürstendamm 100",
					City:    "Berlin",
					Country: "Germany",
				},
			}, nil
		},
	}

	svc := New(nil, client)
	mux := http.NewServeMux()
	path, handler := serviceconnect.NewDeliveryHandler(svc)
	mux.Handle(path, handler)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	deliveryClient := serviceconnect.NewDeliveryClient(server.Client(), server.URL)
	req := connect.NewRequest(&servicepb.QueryRandomAddressRequest{})

	resp, err := deliveryClient.QueryRandomAddress(context.Background(), req)
	if err != nil {
		t.Fatalf("QueryRandomAddress over Connect returned error: %v", err)
	}

	addr := resp.Msg.GetRandomAddress().GetAddress()
	if addr.GetStreet() != "Kurfürstendamm 100" || addr.GetCity() != "Berlin" {
		t.Fatalf("unexpected address: %+v", addr)
	}
}

func TestQueryDeliveryTrackingMapsCourierAndEta(t *testing.T) {
	t.Parallel()

	client := &stubDeliveryClient{
		getOrderTrackingFunc: func(ctx context.Context, _ *deliverygrpc.GetOrderTrackingRequest, _ ...grpc.CallOption) (*deliverygrpc.GetOrderTrackingResponse, error) {
			return &deliverygrpc.GetOrderTrackingResponse{
				OrderId:                   "ord-1",
				PackageId:                 "pkg-1",
				Status:                    deliverygrpc.PackageStatus_PACKAGE_STATUS_IN_TRANSIT,
				EstimatedMinutesRemaining: ptrInt32(14),
				DistanceKmRemaining:       ptrFloat64(3.4),
				Courier: &deliverygrpc.TrackingCourier{
					CourierId:     "courier-7",
					Name:          "Alex Rider",
					Phone:         "+4912345678",
					TransportType: deliverygrpc.TransportType_TRANSPORT_TYPE_BICYCLE,
					Status:        deliverygrpc.CourierStatus_COURIER_STATUS_BUSY,
					CurrentLocation: &deliverygrpc.Location{
						Latitude:  52.52,
						Longitude: 13.41,
					},
				},
			}, nil
		},
	}

	svc := New(nil, client)
	req := connect.NewRequest(&servicepb.QueryDeliveryTrackingRequest{OrderId: "ord-1"})
	req.Header().Set("X-User-ID", "550e8400-e29b-41d4-a716-446655440000")

	resp, err := svc.QueryDeliveryTracking(context.Background(), req)
	if err != nil {
		t.Fatalf("QueryDeliveryTracking returned error: %v", err)
	}

	tracking := resp.Msg.GetDeliveryTracking()
	if tracking == nil {
		t.Fatal("expected tracking payload")
	}
	if tracking.GetOrderId() != "ord-1" {
		t.Fatalf("expected order_id ord-1, got %s", tracking.GetOrderId())
	}
	if tracking.GetPackageId() != "pkg-1" {
		t.Fatalf("expected package_id pkg-1, got %s", tracking.GetPackageId())
	}
	if tracking.GetStatus() != "IN_TRANSIT" {
		t.Fatalf("expected status IN_TRANSIT, got %s", tracking.GetStatus())
	}
	if tracking.GetEstimatedMinutesRemaining() != 14 {
		t.Fatalf("expected eta 14, got %d", tracking.GetEstimatedMinutesRemaining())
	}
	if tracking.GetCourier() == nil || tracking.GetCourier().GetCourierId() != "courier-7" {
		t.Fatalf("expected courier courier-7, got %+v", tracking.GetCourier())
	}
	if tracking.GetCourier().GetTransportType() != "BICYCLE" {
		t.Fatalf("expected transport BICYCLE, got %s", tracking.GetCourier().GetTransportType())
	}
}

func TestQueryDeliveryTrackingReturnsNilOnNotFound(t *testing.T) {
	t.Parallel()

	client := &stubDeliveryClient{
		getOrderTrackingFunc: func(ctx context.Context, _ *deliverygrpc.GetOrderTrackingRequest, _ ...grpc.CallOption) (*deliverygrpc.GetOrderTrackingResponse, error) {
			return nil, grpcstatus.Error(grpccodes.NotFound, "tracking not found")
		},
	}

	svc := New(nil, client)
	req := connect.NewRequest(&servicepb.QueryDeliveryTrackingRequest{OrderId: "ord-missing"})
	req.Header().Set("X-User-ID", "550e8400-e29b-41d4-a716-446655440000")

	resp, err := svc.QueryDeliveryTracking(context.Background(), req)
	if err != nil {
		t.Fatalf("expected nil error for not found, got %v", err)
	}
	if resp.Msg.GetDeliveryTracking() != nil {
		t.Fatalf("expected nil tracking, got %+v", resp.Msg.GetDeliveryTracking())
	}
}

func TestEnvOrDefault(t *testing.T) {
	t.Run("returns value when set", func(t *testing.T) {
		t.Setenv("TEST_KEY", "custom")
		if got := envOrDefault("TEST_KEY", "default"); got != "custom" {
			t.Fatalf("expected custom, got %s", got)
		}
	})

	t.Run("returns fallback when empty", func(t *testing.T) {
		t.Setenv("TEST_KEY_EMPTY", "")
		if got := envOrDefault("TEST_KEY_EMPTY", "fallback"); got != "fallback" {
			t.Fatalf("expected fallback, got %s", got)
		}
	})

	t.Run("returns fallback when unset", func(t *testing.T) {
		os.Unsetenv("TEST_KEY_UNSET")
		if got := envOrDefault("TEST_KEY_UNSET", "default"); got != "default" {
			t.Fatalf("expected default, got %s", got)
		}
	})

	t.Run("returns fallback when whitespace only", func(t *testing.T) {
		t.Setenv("TEST_KEY_WS", "   ")
		if got := envOrDefault("TEST_KEY_WS", "fallback"); got != "fallback" {
			t.Fatalf("expected fallback, got %q", got)
		}
	})
}

func TestNormalizeGRPCTarget(t *testing.T) {
	t.Parallel()

	t.Run("empty returns error", func(t *testing.T) {
		_, err := normalizeGRPCTarget("")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, errEmptyGRPCTarget) {
			t.Fatalf("expected errEmptyGRPCTarget, got %v", err)
		}
	})

	t.Run("host:port returns as-is", func(t *testing.T) {
		got, err := normalizeGRPCTarget("localhost:50052")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "localhost:50052" {
			t.Fatalf("expected localhost:50052, got %s", got)
		}
	})

	t.Run("http URL extracts host:port", func(t *testing.T) {
		got, err := normalizeGRPCTarget("http://shortlink-shop-delivery:50051")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "shortlink-shop-delivery:50051" {
			t.Fatalf("expected shortlink-shop-delivery:50051, got %s", got)
		}
	})

	t.Run("URL without host returns error", func(t *testing.T) {
		_, err := normalizeGRPCTarget("http://")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, errMissingHost) {
			t.Fatalf("expected errMissingHost, got %v", err)
		}
	})
}

// stubDeliveryClient implements deliverygrpc.DeliveryServiceClient for tests.
type stubDeliveryClient struct {
	getRandomAddressFunc       func(context.Context, *deliverygrpc.GetRandomAddressRequest, ...grpc.CallOption) (*deliverygrpc.GetRandomAddressResponse, error)
	getOrderTrackingFunc       func(context.Context, *deliverygrpc.GetOrderTrackingRequest, ...grpc.CallOption) (*deliverygrpc.GetOrderTrackingResponse, error)
	subscribeOrderTrackingFunc func(context.Context, *deliverygrpc.GetOrderTrackingRequest, ...grpc.CallOption) (grpc.ServerStreamingClient[deliverygrpc.GetOrderTrackingResponse], error)
}

func (s *stubDeliveryClient) GetRandomAddress(
	ctx context.Context,
	in *deliverygrpc.GetRandomAddressRequest,
	opts ...grpc.CallOption,
) (*deliverygrpc.GetRandomAddressResponse, error) {
	if s.getRandomAddressFunc != nil {
		return s.getRandomAddressFunc(ctx, in, opts...)
	}
	return &deliverygrpc.GetRandomAddressResponse{}, nil
}

func (s *stubDeliveryClient) GetOrderTracking(
	ctx context.Context,
	in *deliverygrpc.GetOrderTrackingRequest,
	opts ...grpc.CallOption,
) (*deliverygrpc.GetOrderTrackingResponse, error) {
	if s.getOrderTrackingFunc != nil {
		return s.getOrderTrackingFunc(ctx, in, opts...)
	}
	return &deliverygrpc.GetOrderTrackingResponse{}, nil
}

func (s *stubDeliveryClient) SubscribeOrderTracking(
	ctx context.Context,
	in *deliverygrpc.GetOrderTrackingRequest,
	opts ...grpc.CallOption,
) (grpc.ServerStreamingClient[deliverygrpc.GetOrderTrackingResponse], error) {
	if s.subscribeOrderTrackingFunc != nil {
		return s.subscribeOrderTrackingFunc(ctx, in, opts...)
	}
	return &stubTrackingStream{ctx: ctx}, nil
}

type stubTrackingStream struct {
	ctx       context.Context
	responses []*deliverygrpc.GetOrderTrackingResponse
	index     int
}

func (s *stubTrackingStream) Header() (metadata.MD, error) { return metadata.MD{}, nil }
func (s *stubTrackingStream) Trailer() metadata.MD         { return metadata.MD{} }
func (s *stubTrackingStream) CloseSend() error             { return nil }
func (s *stubTrackingStream) Context() context.Context {
	if s.ctx != nil {
		return s.ctx
	}
	return context.Background()
}
func (s *stubTrackingStream) SendMsg(any) error { return nil }
func (s *stubTrackingStream) RecvMsg(any) error { return nil }
func (s *stubTrackingStream) Recv() (*deliverygrpc.GetOrderTrackingResponse, error) {
	if s.index >= len(s.responses) {
		return nil, io.EOF
	}
	resp := s.responses[s.index]
	s.index++
	return resp, nil
}

func ptrInt32(v int32) *int32 {
	return &v
}

func ptrFloat64(v float64) *float64 {
	return &v
}
