package service //nolint:testpackage // testing exported API only

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
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
					Street:     "Alexanderplatz 1",
					City:       "Berlin",
					PostalCode: "10178",
					Country:    "Germany",
					Latitude:   52.5219,
					Longitude:  13.4132,
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
	if addr.GetPostalCode() != "10178" {
		t.Fatalf("expected postal_code 10178, got %s", addr.GetPostalCode())
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
	getRandomAddressFunc func(context.Context, *deliverygrpc.GetRandomAddressRequest, ...grpc.CallOption) (*deliverygrpc.GetRandomAddressResponse, error)
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

func (s *stubDeliveryClient) GetCourier(context.Context, *deliverygrpc.GetCourierRequest, ...grpc.CallOption) (*deliverygrpc.GetCourierResponse, error) {
	return nil, nil
}
func (s *stubDeliveryClient) GetCourierPool(context.Context, *deliverygrpc.GetCourierPoolRequest, ...grpc.CallOption) (*deliverygrpc.GetCourierPoolResponse, error) {
	return nil, nil
}
func (s *stubDeliveryClient) RegisterCourier(context.Context, *deliverygrpc.RegisterCourierRequest, ...grpc.CallOption) (*deliverygrpc.RegisterCourierResponse, error) {
	return nil, nil
}
func (s *stubDeliveryClient) ActivateCourier(context.Context, *deliverygrpc.ActivateCourierRequest, ...grpc.CallOption) (*deliverygrpc.ActivateCourierResponse, error) {
	return nil, nil
}
func (s *stubDeliveryClient) DeactivateCourier(context.Context, *deliverygrpc.DeactivateCourierRequest, ...grpc.CallOption) (*deliverygrpc.DeactivateCourierResponse, error) {
	return nil, nil
}
func (s *stubDeliveryClient) ArchiveCourier(context.Context, *deliverygrpc.ArchiveCourierRequest, ...grpc.CallOption) (*deliverygrpc.ArchiveCourierResponse, error) {
	return nil, nil
}
func (s *stubDeliveryClient) UpdateContactInfo(context.Context, *deliverygrpc.UpdateContactInfoRequest, ...grpc.CallOption) (*deliverygrpc.UpdateContactInfoResponse, error) {
	return nil, nil
}
func (s *stubDeliveryClient) UpdateWorkSchedule(context.Context, *deliverygrpc.UpdateWorkScheduleRequest, ...grpc.CallOption) (*deliverygrpc.UpdateWorkScheduleResponse, error) {
	return nil, nil
}
func (s *stubDeliveryClient) ChangeTransportType(context.Context, *deliverygrpc.ChangeTransportTypeRequest, ...grpc.CallOption) (*deliverygrpc.ChangeTransportTypeResponse, error) {
	return nil, nil
}
func (s *stubDeliveryClient) AcceptOrder(context.Context, *deliverygrpc.AcceptOrderRequest, ...grpc.CallOption) (*deliverygrpc.AcceptOrderResponse, error) {
	return nil, nil
}
func (s *stubDeliveryClient) AssignOrder(context.Context, *deliverygrpc.AssignOrderRequest, ...grpc.CallOption) (*deliverygrpc.AssignOrderResponse, error) {
	return nil, nil
}
func (s *stubDeliveryClient) PickUpOrder(context.Context, *deliverygrpc.PickUpOrderRequest, ...grpc.CallOption) (*deliverygrpc.PickUpOrderResponse, error) {
	return nil, nil
}
func (s *stubDeliveryClient) DeliverOrder(context.Context, *deliverygrpc.DeliverOrderRequest, ...grpc.CallOption) (*deliverygrpc.DeliverOrderResponse, error) {
	return nil, nil
}
func (s *stubDeliveryClient) GetCourierDeliveries(context.Context, *deliverygrpc.GetCourierDeliveriesRequest, ...grpc.CallOption) (*deliverygrpc.GetCourierDeliveriesResponse, error) {
	return nil, nil
}
