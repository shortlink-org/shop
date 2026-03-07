package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
	serviceconnect "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1/v1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	cartgrpc "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/cart/v1"
	cartmodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/cart/v1/model/v1"
	ordergrpc "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/order/v1"
	ordermodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/order/v1/model/v1"
)

const (
	defaultListenAddr = "0.0.0.0:4011"
	defaultOMSGRPCURL = "http://localhost:50051"
	userIDHeader      = "X-User-ID"
	authHeader        = "Authorization"
)

type Service struct {
	logger      *slog.Logger
	cartClient  cartgrpc.CartServiceClient
	orderClient ordergrpc.OrderServiceClient
}

func New(logger *slog.Logger, cartClient cartgrpc.CartServiceClient, orderClient ordergrpc.OrderServiceClient) *Service {
	if logger == nil {
		logger = slog.Default()
	}

	return &Service{
		logger:      logger,
		cartClient:  cartClient,
		orderClient: orderClient,
	}
}

func Start(ctx context.Context) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	listenAddr := envOrDefault("LISTEN_ADDR", defaultListenAddr)
	omsTarget, err := normalizeGRPCTarget(envOrDefault("OMS_GRPC_URL", defaultOMSGRPCURL))
	if err != nil {
		return fmt.Errorf("normalize OMS_GRPC_URL: %w", err)
	}

	conn, err := grpc.NewClient(
		omsTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("create oms grpc client: %w", err)
	}
	defer conn.Close()

	svc := New(logger, cartgrpc.NewCartServiceClient(conn), ordergrpc.NewOrderServiceClient(conn))

	mux := http.NewServeMux()
	path, handler := serviceconnect.NewShopHandler(svc)
	mux.Handle(path, handler)
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    listenAddr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	logger.Info("starting oms-graphql grpc subgraph", "listen_addr", listenAddr, "oms_target", omsTarget)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown oms-graphql grpc subgraph", "error", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Service) QueryGetCart(
	ctx context.Context,
	req *connect.Request[servicepb.QueryGetCartRequest],
) (*connect.Response[servicepb.QueryGetCartResponse], error) {
	userID, outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	response, err := s.cartClient.Get(outboundCtx, &cartmodel.GetRequest{
		CustomerId: userID,
	})
	if err != nil {
		return nil, grpcError(err)
	}

	return connect.NewResponse(&servicepb.QueryGetCartResponse{
		GetCart: &servicepb.GetCartResponse{
			State: mapCartState(response.GetState()),
		},
	}), nil
}

func (s *Service) QueryGetOrder(
	ctx context.Context,
	req *connect.Request[servicepb.QueryGetOrderRequest],
) (*connect.Response[servicepb.QueryGetOrderResponse], error) {
	_, outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	response, err := s.orderClient.Get(outboundCtx, &ordermodel.GetRequest{
		Id: req.Msg.GetId(),
	})
	if err != nil {
		return nil, grpcError(err)
	}

	return connect.NewResponse(&servicepb.QueryGetOrderResponse{
		GetOrder: &servicepb.GetOrderResponse{
			Order: mapOrderState(response.GetOrder()),
		},
	}), nil
}

func (s *Service) MutationAddItem(
	ctx context.Context,
	req *connect.Request[servicepb.MutationAddItemRequest],
) (*connect.Response[servicepb.MutationAddItemResponse], error) {
	userID, outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	_, err = s.cartClient.Add(outboundCtx, &cartmodel.AddRequest{
		CustomerId: userID,
		Items:      mapCartItemInputs(req.Msg.GetAddRequest()),
	})
	if err != nil {
		return nil, grpcError(err)
	}

	return connect.NewResponse(&servicepb.MutationAddItemResponse{
		AddItem: okEmpty(),
	}), nil
}

func (s *Service) MutationRemoveItem(
	ctx context.Context,
	req *connect.Request[servicepb.MutationRemoveItemRequest],
) (*connect.Response[servicepb.MutationRemoveItemResponse], error) {
	userID, outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	_, err = s.cartClient.Remove(outboundCtx, &cartmodel.RemoveRequest{
		CustomerId: userID,
		Items:      mapCartItemInputs(req.Msg.GetRemoveRequest()),
	})
	if err != nil {
		return nil, grpcError(err)
	}

	return connect.NewResponse(&servicepb.MutationRemoveItemResponse{
		RemoveItem: okEmpty(),
	}), nil
}

func (s *Service) MutationResetCart(
	ctx context.Context,
	req *connect.Request[servicepb.MutationResetCartRequest],
) (*connect.Response[servicepb.MutationResetCartResponse], error) {
	userID, outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	_, err = s.cartClient.Reset(outboundCtx, &cartmodel.ResetRequest{
		CustomerId: userID,
	})
	if err != nil {
		return nil, grpcError(err)
	}

	return connect.NewResponse(&servicepb.MutationResetCartResponse{
		ResetCart: okEmpty(),
	}), nil
}

func (s *Service) MutationCreateOrder(
	ctx context.Context,
	req *connect.Request[servicepb.MutationCreateOrderRequest],
) (*connect.Response[servicepb.MutationCreateOrderResponse], error) {
	userID, outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	createRequest, err := mapCreateOrderRequest(userID, req.Msg.GetInput())
	if err != nil {
		return nil, err
	}

	_, err = s.orderClient.Create(outboundCtx, createRequest)
	if err != nil {
		return nil, grpcError(err)
	}

	return connect.NewResponse(&servicepb.MutationCreateOrderResponse{
		CreateOrder: okEmpty(),
	}), nil
}

func (s *Service) MutationCancelOrder(
	ctx context.Context,
	req *connect.Request[servicepb.MutationCancelOrderRequest],
) (*connect.Response[servicepb.MutationCancelOrderResponse], error) {
	_, outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	_, err = s.orderClient.Cancel(outboundCtx, &ordermodel.CancelRequest{
		Id: req.Msg.GetOrderId(),
	})
	if err != nil {
		return nil, grpcError(err)
	}

	return connect.NewResponse(&servicepb.MutationCancelOrderResponse{
		CancelOrder: okEmpty(),
	}), nil
}

func (s *Service) MutationUpdateDeliveryInfo(
	ctx context.Context,
	req *connect.Request[servicepb.MutationUpdateDeliveryInfoRequest],
) (*connect.Response[servicepb.MutationUpdateDeliveryInfoResponse], error) {
	_, outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	input := req.Msg.GetInput()
	if input == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("input is required"))
	}

	deliveryInfo, err := mapDeliveryInfoInput(input.GetDeliveryInfo())
	if err != nil {
		return nil, err
	}

	_, err = s.orderClient.UpdateDeliveryInfo(outboundCtx, &ordermodel.UpdateDeliveryInfoRequest{
		OrderId:      input.GetOrderId(),
		DeliveryInfo: deliveryInfo,
	})
	if err != nil {
		return nil, grpcError(err)
	}

	return connect.NewResponse(&servicepb.MutationUpdateDeliveryInfoResponse{
		UpdateDeliveryInfo: okEmpty(),
	}), nil
}

func (s *Service) MutationCheckout(
	ctx context.Context,
	req *connect.Request[servicepb.MutationCheckoutRequest],
) (*connect.Response[servicepb.MutationCheckoutResponse], error) {
	userID, outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	input := req.Msg.GetInput()
	var deliveryInfoInput *servicepb.DeliveryInfoInput
	if input != nil {
		deliveryInfoInput = input.GetDeliveryInfo()
	}

	deliveryInfo, err := mapDeliveryInfoInput(deliveryInfoInput)
	if err != nil {
		return nil, err
	}

	response, err := s.orderClient.Checkout(outboundCtx, &ordermodel.CheckoutRequest{
		CustomerId:   userID,
		DeliveryInfo: deliveryInfo,
	})
	if err != nil {
		return nil, grpcError(err)
	}

	return connect.NewResponse(&servicepb.MutationCheckoutResponse{
		Checkout: &servicepb.CheckoutResult{
			OrderId: wrapperspb.String(response.GetOrderId()),
		},
	}), nil
}

func (s *Service) authorizedContext(
	ctx context.Context,
	req interface{ Header() http.Header },
) (string, context.Context, error) {
	userID := strings.TrimSpace(req.Header().Get(userIDHeader))
	if userID == "" {
		return "", nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing required X-User-ID header"))
	}

	mdPairs := []string{"x-user-id", userID}
	if authorization := strings.TrimSpace(req.Header().Get(authHeader)); authorization != "" {
		mdPairs = append(mdPairs, "authorization", authorization)
	}

	return userID, metadata.NewOutgoingContext(ctx, metadata.Pairs(mdPairs...)), nil
}

func grpcError(err error) error {
	st, ok := grpcstatus.FromError(err)
	if !ok {
		return connect.NewError(connect.CodeUnknown, err)
	}

	return connect.NewError(mapGRPCCode(st.Code()), err)
}

func mapGRPCCode(code grpccodes.Code) connect.Code {
	switch code {
	case grpccodes.Canceled:
		return connect.CodeCanceled
	case grpccodes.Unknown:
		return connect.CodeUnknown
	case grpccodes.InvalidArgument:
		return connect.CodeInvalidArgument
	case grpccodes.DeadlineExceeded:
		return connect.CodeDeadlineExceeded
	case grpccodes.NotFound:
		return connect.CodeNotFound
	case grpccodes.AlreadyExists:
		return connect.CodeAlreadyExists
	case grpccodes.PermissionDenied:
		return connect.CodePermissionDenied
	case grpccodes.ResourceExhausted:
		return connect.CodeResourceExhausted
	case grpccodes.FailedPrecondition:
		return connect.CodeFailedPrecondition
	case grpccodes.Aborted:
		return connect.CodeAborted
	case grpccodes.OutOfRange:
		return connect.CodeOutOfRange
	case grpccodes.Unimplemented:
		return connect.CodeUnimplemented
	case grpccodes.Internal:
		return connect.CodeInternal
	case grpccodes.Unavailable:
		return connect.CodeUnavailable
	case grpccodes.DataLoss:
		return connect.CodeDataLoss
	case grpccodes.Unauthenticated:
		return connect.CodeUnauthenticated
	default:
		return connect.CodeUnknown
	}
}

func envOrDefault(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}

func normalizeGRPCTarget(raw string) (string, error) {
	if raw == "" {
		return "", errors.New("empty grpc target")
	}

	if strings.Contains(raw, "://") {
		parsed, err := url.Parse(raw)
		if err != nil {
			return "", err
		}
		if parsed.Host == "" {
			return "", fmt.Errorf("missing host in %q", raw)
		}

		return parsed.Host, nil
	}

	return raw, nil
}

func okEmpty() *servicepb.Empty {
	return &servicepb.Empty{
		Ok: wrapperspb.Bool(true),
	}
}
