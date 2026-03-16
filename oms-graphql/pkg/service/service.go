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
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/shortlink-org/shop/oms-graphql/pkg/dto"
	cartgrpc "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/cart/v1"
	cartmodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/cart/v1/model/v1"
	ordergrpc "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/order/v1"
	ordermodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/order/v1/model/v1"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
	serviceconnect "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1/v1connect"
)

const (
	defaultListenAddr = "0.0.0.0:4011"
	defaultOMSGRPCURL = "http://localhost:50051"
	authHeader        = "Authorization"
	traceparentHeader = "Traceparent"
	traceIDHeader     = "Trace-Id"
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
	shutdownTracing, err := setupTracing(ctx, "shortlink-shop-oms-graphql")
	if err != nil {
		return fmt.Errorf("setup tracing: %w", err)
	}
	defer func() {
		if shutdownErr := shutdownTracing(context.Background()); shutdownErr != nil {
			logger.Error("shutdown tracing", "error", shutdownErr)
		}
	}()

	omsTarget, err := normalizeGRPCTarget(envOrDefault("OMS_GRPC_URL", defaultOMSGRPCURL))
	if err != nil {
		return fmt.Errorf("normalize OMS_GRPC_URL: %w", err)
	}

	conn, err := grpc.NewClient(
		omsTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return fmt.Errorf("create oms grpc client: %w", err)
	}

	defer func() { _ = conn.Close() }() //nolint:errcheck // best-effort on shutdown

	svc := New(logger, cartgrpc.NewCartServiceClient(conn), ordergrpc.NewOrderServiceClient(conn))

	mux := http.NewServeMux()
	path, handler := serviceconnect.NewShopHandler(svc)
	mux.Handle(path, handler)
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok")) //nolint:errcheck // healthz best-effort
	}))
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              listenAddr,
		Handler:           otelhttp.NewHandler(withRequestLogging(logger, h2c.NewHandler(mux, &http2.Server{})), "oms-graphql"),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("starting oms-graphql grpc subgraph", "listen_addr", listenAddr, "oms_target", omsTarget)

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		shutdownErr := server.Shutdown(shutdownCtx)
		if shutdownErr != nil {
			logger.Error("shutdown oms-graphql grpc subgraph", "error", shutdownErr)
		}
	}()

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

func (s *Service) QueryGetCart(
	ctx context.Context,
	req *connect.Request[servicepb.QueryGetCartRequest],
) (*connect.Response[servicepb.QueryGetCartResponse], error) { //nolint:whitespace // multi-line signature
	outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	response, err := s.cartClient.Get(outboundCtx, &cartmodel.GetRequest{})
	if err != nil {
		return nil, grpcError(err)
	}

	return connect.NewResponse(&servicepb.QueryGetCartResponse{
		GetCart: &servicepb.GetCartResponse{
			State: dto.CartStateToService(response.GetState()),
		},
	}), nil
}

func (s *Service) QueryGetLeaderboard(
	ctx context.Context,
	req *connect.Request[servicepb.QueryGetLeaderboardRequest],
) (*connect.Response[servicepb.QueryGetLeaderboardResponse], error) { //nolint:whitespace // multi-line signature
	outboundCtx := s.forwardContext(ctx, req)

	limit := int32(7)
	if req.Msg.GetLimit() != nil {
		limit = req.Msg.GetLimit().GetValue()
	}

	response, err := s.orderClient.GetLeaderboard(outboundCtx, &ordermodel.GetLeaderboardRequest{
		Board:  req.Msg.GetBoard(),
		Window: req.Msg.GetWindow(),
		Limit:  limit,
	})
	if err != nil {
		return nil, grpcError(err)
	}

	return connect.NewResponse(&servicepb.QueryGetLeaderboardResponse{
		GetLeaderboard: &servicepb.GetLeaderboardResponse{
			Leaderboard: dto.GoodsLeaderboardToService(response.GetLeaderboard()),
		},
	}), nil
}

func (s *Service) QueryGetOrder(
	ctx context.Context,
	req *connect.Request[servicepb.QueryGetOrderRequest],
) (*connect.Response[servicepb.QueryGetOrderResponse], error) { //nolint:whitespace // multi-line signature
	outboundCtx, err := s.authorizedContext(ctx, req)
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
			Order: dto.OrderStateToService(response.GetOrder()),
		},
	}), nil
}

func (s *Service) MutationAddItem(
	ctx context.Context,
	req *connect.Request[servicepb.MutationAddItemRequest],
) (*connect.Response[servicepb.MutationAddItemResponse], error) { //nolint:whitespace // multi-line signature
	outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	_, err = s.cartClient.Add(outboundCtx, &cartmodel.AddRequest{
		Items: dto.CartItemInputsToOMS(req.Msg.GetAddRequest()),
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
) (*connect.Response[servicepb.MutationRemoveItemResponse], error) { //nolint:whitespace // multi-line signature
	outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	_, err = s.cartClient.Remove(outboundCtx, &cartmodel.RemoveRequest{
		Items: dto.CartItemInputsToOMS(req.Msg.GetRemoveRequest()),
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
) (*connect.Response[servicepb.MutationResetCartResponse], error) { //nolint:whitespace // multi-line signature
	outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	_, err = s.cartClient.Reset(outboundCtx, &cartmodel.ResetRequest{})
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
) (*connect.Response[servicepb.MutationCreateOrderResponse], error) { //nolint:whitespace // multi-line signature
	outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	createRequest, err := dto.CreateOrderRequestFromInput(req.Msg.GetInput())
	if err != nil {
		return nil, fmt.Errorf("create order request: %w", err)
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
) (*connect.Response[servicepb.MutationCancelOrderResponse], error) { //nolint:whitespace // multi-line signature
	outboundCtx, err := s.authorizedContext(ctx, req)
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
) (*connect.Response[servicepb.MutationUpdateDeliveryInfoResponse], error) { //nolint:whitespace // multi-line signature
	outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	input := req.Msg.GetInput()
	if input == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errInputRequired)
	}

	deliveryInfo, err := dto.DeliveryInfoFromInput(input.GetDeliveryInfo())
	if err != nil {
		return nil, fmt.Errorf("delivery info: %w", err)
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
) (*connect.Response[servicepb.MutationCheckoutResponse], error) { //nolint:whitespace // multi-line signature
	outboundCtx, err := s.authorizedContext(ctx, req)
	if err != nil {
		return nil, err
	}

	input := req.Msg.GetInput()

	var deliveryInfoInput *servicepb.DeliveryInfoInput
	if input != nil {
		deliveryInfoInput = input.GetDeliveryInfo()
	}

	deliveryInfo, err := dto.DeliveryInfoFromInput(deliveryInfoInput)
	if err != nil {
		return nil, fmt.Errorf("delivery info: %w", err)
	}

	response, err := s.orderClient.Checkout(outboundCtx, &ordermodel.CheckoutRequest{
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

var (
	errMissingAuthorization = errors.New("missing required Authorization header")
	errInputRequired        = errors.New("input is required")
	errEmptyGRPCTarget      = errors.New("empty grpc target")
	errMissingHost          = errors.New("missing host in URL")
)

func (s *Service) authorizedContext(
	ctx context.Context,
	req interface{ Header() http.Header },
) (context.Context, error) { //nolint:whitespace // multi-line signature
	authorization := strings.TrimSpace(req.Header().Get(authHeader))
	if authorization == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errMissingAuthorization)
	}

	return s.forwardContext(ctx, req), nil
}

func (s *Service) forwardContext(
	ctx context.Context,
	req interface{ Header() http.Header },
) context.Context { //nolint:whitespace // multi-line signature
	authorization := strings.TrimSpace(req.Header().Get(authHeader))

	mdPairs := make([]string, 0, 6)
	if authorization != "" {
		mdPairs = append(mdPairs, "authorization", authorization)
	}

	if v := strings.TrimSpace(req.Header().Get(traceparentHeader)); v != "" {
		mdPairs = append(mdPairs, "traceparent", v)
	}

	if v := strings.TrimSpace(req.Header().Get(traceIDHeader)); v != "" {
		mdPairs = append(mdPairs, "trace-id", v)
	}

	if len(mdPairs) == 0 {
		return ctx
	}

	return metadata.NewOutgoingContext(ctx, metadata.Pairs(mdPairs...))
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

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}

func normalizeGRPCTarget(raw string) (string, error) {
	if raw == "" {
		return "", errEmptyGRPCTarget
	}

	if strings.Contains(raw, "://") {
		parsed, err := url.Parse(raw)
		if err != nil {
			return "", fmt.Errorf("parse grpc target: %w", err)
		}

		if parsed.Host == "" {
			return "", fmt.Errorf("%w: %q", errMissingHost, raw)
		}

		return parsed.Host, nil
	}

	return raw, nil
}

func withRequestLogging(logger *slog.Logger, next http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(recorder, r)

		logger.Info(
			"oms-graphql request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.statusCode,
			"duration_ms", time.Since(startedAt).Milliseconds(),
			"trace_id", traceIDFromRequest(r),
			"request_trace_id", strings.TrimSpace(r.Header.Get(traceIDHeader)),
			"has_traceparent", strings.TrimSpace(r.Header.Get(traceparentHeader)) != "",
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func traceIDFromRequest(r *http.Request) string {
	if spanCtx := oteltrace.SpanContextFromContext(r.Context()); spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}

	if traceID := strings.TrimSpace(r.Header.Get(traceIDHeader)); traceID != "" {
		return traceID
	}

	if traceparent := strings.TrimSpace(r.Header.Get(traceparentHeader)); traceparent != "" {
		parts := strings.Split(traceparent, "-")
		if len(parts) >= 2 && len(parts[1]) == 32 {
			return parts[1]
		}
	}

	return ""
}

func okEmpty() *servicepb.Empty {
	return &servicepb.Empty{
		Ok: wrapperspb.Bool(true),
	}
}

func setupTracing(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	if strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")) == "" {
		return func(context.Context) error { return nil }, nil
	}

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create OTLP HTTP exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("build tracing resource: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return provider.Shutdown, nil
}
