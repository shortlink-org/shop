package service //nolint:testpackage // testing exported API only

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/shortlink-org/shop/oms-graphql/pkg/dto"
	commonpb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/domain/order/v1/common"
	cartgrpc "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/cart/v1"
	cartmodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/cart/v1/model/v1"
	ordergrpc "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/order/v1"
	ordermodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/order/v1/model/v1"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
	serviceconnect "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1/v1connect"
)

const (
	testOrderID1   = "order-1"
	testYear       = 2026
	testWeightKg   = 3.5
	testMarch12_9h = 12
)

func TestQueryGetCartRequiresUserID(t *testing.T) {
	t.Parallel()

	svc := New(nil, nil, nil)
	req := connect.NewRequest(&servicepb.QueryGetCartRequest{})

	_, err := svc.QueryGetCart(context.Background(), req)
	if err == nil {
		t.Fatal("expected unauthenticated error")
	}

	if code := connect.CodeOf(err); code != connect.CodeUnauthenticated {
		t.Fatalf("expected code %v, got %v", connect.CodeUnauthenticated, code)
	}
}

func TestMapCreateOrderRequestGeneratesOrderID(t *testing.T) {
	t.Parallel()

	req, err := dto.CreateOrderRequestFromInput(&servicepb.CreateOrderInput{
		Items: []*servicepb.OrderItemInput{
			{
				Id:       "good-1",
				Quantity: 2,
				Price:    99.5,
			},
		},
	})
	if err != nil {
		t.Fatalf("mapCreateOrderRequest returned error: %v", err)
	}

	if req.GetOrder().GetId() == "" {
		t.Fatal("expected generated order id")
	}

	if len(req.GetOrder().GetItems()) != 1 {
		t.Fatalf("expected 1 order item, got %d", len(req.GetOrder().GetItems()))
	}
}

func TestQueryGetOrderMapsDeliveryLifecycleFields(t *testing.T) {
	t.Parallel()

	requestedAt := time.Date(2026, time.March, 11, 14, 30, 0, 0, time.UTC)
	client := &stubOrderClient{
		getFunc: func(ctx context.Context, in *ordermodel.GetRequest, _ ...grpc.CallOption) (*ordermodel.GetResponse, error) {
			if in.GetId() != "order-1" {
				t.Fatalf("expected order id order-1, got %s", in.GetId())
			}

			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				t.Fatal("expected outgoing metadata")
			}

			if got := md.Get("authorization"); len(got) != 1 || got[0] != "Bearer test-token" {
				t.Fatalf("expected authorization metadata, got %v", got)
			}

			return &ordermodel.GetResponse{
				Order: &ordermodel.OrderState{
					Id:             testOrderID1,
					Status:         commonpb.OrderStatus_ORDER_STATUS_PROCESSING,
					DeliveryStatus: commonpb.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT,
					PackageId:      "package-1",
					RequestedAt:    timestamppb.New(requestedAt),
				},
			}, nil
		},
	}

	svc := New(nil, nil, client)
	req := connect.NewRequest(&servicepb.QueryGetOrderRequest{Id: testOrderID1})
	req.Header().Set(authHeader, "Bearer test-token")

	resp, err := svc.QueryGetOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("QueryGetOrder returned error: %v", err)
	}

	order := resp.Msg.GetGetOrder().GetOrder()
	if order.GetId().GetValue() != testOrderID1 {
		t.Fatalf("expected order id %s, got %s", testOrderID1, order.GetId().GetValue())
	}

	if order.GetDeliveryStatus().GetValue() != "DELIVERY_STATUS_IN_TRANSIT" {
		t.Fatalf("expected delivery status DELIVERY_STATUS_IN_TRANSIT, got %s", order.GetDeliveryStatus().GetValue())
	}

	if order.GetPackageId().GetValue() != "package-1" {
		t.Fatalf("expected package id package-1, got %s", order.GetPackageId().GetValue())
	}

	if !order.GetRequestedAt().AsTime().Equal(requestedAt) {
		t.Fatalf("expected requested_at %s, got %s", requestedAt, order.GetRequestedAt().AsTime())
	}
}

func TestShopHandlerQueryGetOrderMapsDeliveryLifecycleFields(t *testing.T) {
	t.Parallel()

	requestedAt := time.Date(testYear, time.March, 11, 15, 45, 0, 0, time.UTC)
	client := &stubOrderClient{
		getFunc: func(ctx context.Context, in *ordermodel.GetRequest, _ ...grpc.CallOption) (*ordermodel.GetResponse, error) {
			if in.GetId() != "order-connect" {
				t.Fatalf("expected order id order-connect, got %s", in.GetId())
			}

			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				t.Fatal("expected outgoing metadata")
			}

			if got := md.Get("authorization"); len(got) != 1 || got[0] != "Bearer connect-token" {
				t.Fatalf("expected authorization metadata, got %v", got)
			}

			return &ordermodel.GetResponse{
				Order: &ordermodel.OrderState{
					Id:             "order-connect",
					Status:         commonpb.OrderStatus_ORDER_STATUS_PROCESSING,
					DeliveryStatus: commonpb.DeliveryStatus_DELIVERY_STATUS_ACCEPTED,
					PackageId:      "package-connect",
					RequestedAt:    timestamppb.New(requestedAt),
				},
			}, nil
		},
	}

	svc := New(nil, nil, client)
	mux := http.NewServeMux()
	path, handler := serviceconnect.NewShopHandler(svc)
	mux.Handle(path, handler)

	server := httptest.NewServer(mux)
	defer server.Close()

	shopClient := serviceconnect.NewShopClient(server.Client(), server.URL)
	req := connect.NewRequest(&servicepb.QueryGetOrderRequest{Id: "order-connect"})
	req.Header().Set(authHeader, "Bearer connect-token")

	resp, err := shopClient.QueryGetOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("QueryGetOrder over Connect returned error: %v", err)
	}

	order := resp.Msg.GetGetOrder().GetOrder()
	if order.GetId().GetValue() != "order-connect" {
		t.Fatalf("expected order id order-connect, got %s", order.GetId().GetValue())
	}

	if order.GetDeliveryStatus().GetValue() != "DELIVERY_STATUS_ACCEPTED" {
		t.Fatalf("expected delivery status DELIVERY_STATUS_ACCEPTED, got %s", order.GetDeliveryStatus().GetValue())
	}

	if order.GetPackageId().GetValue() != "package-connect" {
		t.Fatalf("expected package id package-connect, got %s", order.GetPackageId().GetValue())
	}

	if !order.GetRequestedAt().AsTime().Equal(requestedAt) {
		t.Fatalf("expected requested_at %s, got %s", requestedAt, order.GetRequestedAt().AsTime())
	}
}

func TestShopHandlerQueryGetCartMapsState(t *testing.T) {
	t.Parallel()

	cartClient := &stubCartClient{
		getFunc: func(ctx context.Context, _ *cartmodel.GetRequest, _ ...grpc.CallOption) (*cartmodel.GetResponse, error) {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				t.Fatal("expected outgoing metadata")
			}

			if got := md.Get("authorization"); len(got) != 1 || got[0] != "Bearer cart-token" {
				t.Fatalf("expected authorization metadata, got %v", got)
			}

			return &cartmodel.GetResponse{
				State: &cartmodel.CartState{
					CartId: "cart-1",
					Items: []*cartmodel.CartItem{
						{GoodId: "good-1", Quantity: 2},
						{GoodId: "good-2", Quantity: 1},
					},
				},
			}, nil
		},
	}

	shopClient := newTestShopClient(t, New(nil, cartClient, nil))
	req := connect.NewRequest(&servicepb.QueryGetCartRequest{})
	req.Header().Set(authHeader, "Bearer cart-token")

	resp, err := shopClient.QueryGetCart(context.Background(), req)
	if err != nil {
		t.Fatalf("QueryGetCart over Connect returned error: %v", err)
	}

	state := resp.Msg.GetGetCart().GetState()
	if state.GetCartId().GetValue() != "cart-1" {
		t.Fatalf("expected cart id cart-1, got %s", state.GetCartId().GetValue())
	}

	items := state.GetItems().GetList().GetItems()
	if len(items) != 2 {
		t.Fatalf("expected 2 cart items, got %d", len(items))
	}

	if items[0].GetGoodId().GetValue() != "good-1" || items[0].GetQuantity().GetValue() != 2 {
		t.Fatalf("unexpected first cart item: %+v", items[0])
	}
}

func TestShopHandlerMutationUpdateDeliveryInfoMapsInput(t *testing.T) {
	t.Parallel()

	orderClient := &stubOrderClient{
		updateDeliveryInfoFunc: func(
			ctx context.Context,
			in *ordermodel.UpdateDeliveryInfoRequest,
			_ ...grpc.CallOption,
		) (*emptypb.Empty, error) {

			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				t.Fatal("expected outgoing metadata")
			}

			if got := md.Get("authorization"); len(got) != 1 || got[0] != "Bearer update-token" {
				t.Fatalf("expected authorization metadata, got %v", got)
			}

			if in.GetOrderId() != "order-1" {
				t.Fatalf("expected order id order-1, got %s", in.GetOrderId())
			}

			info := in.GetDeliveryInfo()
			if info == nil {
				t.Fatal("expected delivery info")
			}

			if info.GetPickupAddress().GetStreet() != "Pickup st" {
				t.Fatalf("expected pickup street Pickup st, got %s", info.GetPickupAddress().GetStreet())
			}

			if info.GetDeliveryAddress().GetStreet() != "Delivery st" {
				t.Fatalf("expected delivery street Delivery st, got %s", info.GetDeliveryAddress().GetStreet())
			}

			if info.GetPriority() != commonpb.DeliveryPriority_DELIVERY_PRIORITY_URGENT {
				t.Fatalf("expected urgent priority, got %s", info.GetPriority().String())
			}

			if info.GetPackageInfo().GetWeightKg() != testWeightKg {
				t.Fatalf("expected weight %v, got %v", testWeightKg, info.GetPackageInfo().GetWeightKg())
			}

			if got := info.GetDeliveryPeriod().GetStartTime().AsTime().UTC(); !got.Equal(time.Date(testYear, time.March, testMarch12_9h, 9, 0, 0, 0, time.UTC)) {
				t.Fatalf("unexpected delivery period start: %s", got)
			}

			return &emptypb.Empty{}, nil
		},
	}

	shopClient := newTestShopClient(t, New(nil, nil, orderClient))
	req := connect.NewRequest(&servicepb.MutationUpdateDeliveryInfoRequest{
		Input: &servicepb.UpdateDeliveryInfoInput{
			OrderId: "order-1",
			DeliveryInfo: &servicepb.DeliveryInfoInput{
				PickupAddress: &servicepb.DeliveryAddressInput{
					Street:  "Pickup st",
					City:    "Moscow",
					Country: "Russia",
				},
				DeliveryAddress: &servicepb.DeliveryAddressInput{
					Street:  "Delivery st",
					City:    "Saint Petersburg",
					Country: "Russia",
				},
				DeliveryPeriod: &servicepb.DeliveryPeriodInput{
					StartTime: "2026-03-12T09:00:00Z",
					EndTime:   "2026-03-12T12:00:00Z",
				},
				PackageInfo: &servicepb.PackageInfoInput{
					WeightKg: 3.5,
				},
				Priority: wrapperspb.String("URGENT"),
			},
		},
	})
	req.Header().Set(authHeader, "Bearer update-token")

	resp, err := shopClient.MutationUpdateDeliveryInfo(context.Background(), req)
	if err != nil {
		t.Fatalf("MutationUpdateDeliveryInfo over Connect returned error: %v", err)
	}

	if !resp.Msg.GetUpdateDeliveryInfo().GetOk().GetValue() {
		t.Fatal("expected ok=true response")
	}
}

func newTestShopClient(t *testing.T, svc *Service) serviceconnect.ShopClient { //nolint:ireturn // test helper returns interface for Connect client
	t.Helper()

	mux := http.NewServeMux()
	path, handler := serviceconnect.NewShopHandler(svc)
	mux.Handle(path, handler)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return serviceconnect.NewShopClient(server.Client(), server.URL)
}

type stubOrderClient struct {
	ordergrpc.OrderServiceClient

	getFunc                func(ctx context.Context, in *ordermodel.GetRequest, opts ...grpc.CallOption) (*ordermodel.GetResponse, error)
	updateDeliveryInfoFunc func(ctx context.Context, in *ordermodel.UpdateDeliveryInfoRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

func (s *stubOrderClient) Get(
	ctx context.Context,
	in *ordermodel.GetRequest,
	opts ...grpc.CallOption,
) (*ordermodel.GetResponse, error) {

	if s.getFunc == nil {
		return nil, nil
	}

	return s.getFunc(ctx, in, opts...)
}

func (s *stubOrderClient) Create(context.Context, *ordermodel.CreateRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *stubOrderClient) List(context.Context, *ordermodel.ListRequest, ...grpc.CallOption) (*ordermodel.ListResponse, error) {
	return nil, nil
}

func (s *stubOrderClient) Cancel(context.Context, *ordermodel.CancelRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *stubOrderClient) UpdateDeliveryInfo(
	ctx context.Context,
	in *ordermodel.UpdateDeliveryInfoRequest,
	opts ...grpc.CallOption,
) (*emptypb.Empty, error) {

	if s.updateDeliveryInfoFunc != nil {
		return s.updateDeliveryInfoFunc(ctx, in, opts...)
	}

	return nil, nil
}

func (s *stubOrderClient) Checkout(
	context.Context,
	*ordermodel.CheckoutRequest,
	...grpc.CallOption,
) (*ordermodel.CheckoutResponse, error) {
	return nil, nil
}

type stubCartClient struct {
	cartgrpc.CartServiceClient

	getFunc func(ctx context.Context, in *cartmodel.GetRequest, opts ...grpc.CallOption) (*cartmodel.GetResponse, error)
}

func (s *stubCartClient) Add(context.Context, *cartmodel.AddRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *stubCartClient) Remove(context.Context, *cartmodel.RemoveRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *stubCartClient) Get(
	ctx context.Context,
	in *cartmodel.GetRequest,
	opts ...grpc.CallOption,
) (*cartmodel.GetResponse, error) {
	if s.getFunc != nil {
		return s.getFunc(ctx, in, opts...)
	}

	return nil, nil
}

func (s *stubCartClient) Reset(context.Context, *cartmodel.ResetRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}
