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
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	grpcstatus "google.golang.org/grpc/status"

	deliverygrpc "github.com/shortlink-org/shop/delivery-graphql/pkg/generated/infrastructure/rpc/delivery/v1"
	servicepb "github.com/shortlink-org/shop/delivery-graphql/pkg/generated/service/v1"
	serviceconnect "github.com/shortlink-org/shop/delivery-graphql/pkg/generated/service/v1/v1connect"
)

const (
	defaultListenAddr   = "0.0.0.0:4013"
	defaultDeliveryGRPC = "localhost:50052"
)

var (
	errEmptyGRPCTarget = errors.New("empty gRPC target")
	errMissingHost     = errors.New("missing host in gRPC URL")
)

type Service struct {
	logger         *slog.Logger
	deliveryClient deliverygrpc.DeliveryServiceClient
}

func New(logger *slog.Logger, deliveryClient deliverygrpc.DeliveryServiceClient) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		logger:         logger,
		deliveryClient: deliveryClient,
	}
}

func Start(ctx context.Context) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	listenAddr := envOrDefault("LISTEN_ADDR", defaultListenAddr)
	shutdownTracing, err := setupTracing(ctx, "shortlink-shop-delivery-graphql")
	if err != nil {
		return fmt.Errorf("setup tracing: %w", err)
	}
	defer func() {
		if shutdownErr := shutdownTracing(context.Background()); shutdownErr != nil {
			logger.Error("shutdown tracing", "error", shutdownErr)
		}
	}()

	deliveryTarget, err := normalizeGRPCTarget(envOrDefault("DELIVERY_GRPC_URL", defaultDeliveryGRPC))
	if err != nil {
		return fmt.Errorf("normalize DELIVERY_GRPC_URL: %w", err)
	}

	conn, err := grpc.NewClient(
		deliveryTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return fmt.Errorf("create delivery grpc client: %w", err)
	}

	defer func() { _ = conn.Close() }() //nolint:errcheck // best-effort on shutdown

	svc := New(logger, deliverygrpc.NewDeliveryServiceClient(conn))

	mux := http.NewServeMux()
	path, handler := serviceconnect.NewDeliveryHandler(svc)
	mux.Handle(path, handler)
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok")) //nolint:errcheck // healthz best-effort
	}))
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              listenAddr,
		Handler:           otelhttp.NewHandler(h2c.NewHandler(mux, &http2.Server{}), "delivery-graphql"),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("starting delivery-graphql grpc subgraph", "listen_addr", listenAddr, "delivery_target", deliveryTarget)

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		shutdownErr := server.Shutdown(shutdownCtx)
		if shutdownErr != nil {
			logger.Error("shutdown delivery-graphql grpc subgraph", "error", shutdownErr)
		}
	}()

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

func (s *Service) QueryRandomAddress(
	ctx context.Context,
	req *connect.Request[servicepb.QueryRandomAddressRequest],
) (*connect.Response[servicepb.QueryRandomAddressResponse], error) {
	resp, err := s.getRandomAddress(ctx, req.Header())
	if err != nil {
		return nil, grpcError(err)
	}

	addr := resp.GetAddress()
	if addr == nil {
		return connect.NewResponse(&servicepb.QueryRandomAddressResponse{
			RandomAddress: &servicepb.RandomAddressResponse{},
		}), nil
	}

	return connect.NewResponse(&servicepb.QueryRandomAddressResponse{
		RandomAddress: &servicepb.RandomAddressResponse{
			Address: &servicepb.Address{
				Street:    addr.GetStreet(),
				City:      addr.GetCity(),
				Country:   addr.GetCountry(),
				Latitude:  addr.GetLatitude(),
				Longitude: addr.GetLongitude(),
			},
		},
	}), nil
}

func (s *Service) QueryDeliveryTracking(
	ctx context.Context,
	req *connect.Request[servicepb.QueryDeliveryTrackingRequest],
) (*connect.Response[servicepb.QueryDeliveryTrackingResponse], error) {
	resp, err := s.getOrderTracking(ctx, req.Header(), req.Msg.GetOrderId())
	if err != nil {
		return nil, grpcError(err)
	}
	if resp == nil {
		return connect.NewResponse(&servicepb.QueryDeliveryTrackingResponse{}), nil
	}

	return connect.NewResponse(&servicepb.QueryDeliveryTrackingResponse{
		DeliveryTracking: deliveryTrackingToService(resp),
	}), nil
}

func deliveryTrackingToService(resp *deliverygrpc.GetOrderTrackingResponse) *servicepb.DeliveryTracking {
	if resp == nil {
		return nil
	}

	var courier *servicepb.DeliveryCourier
	if source := resp.GetCourier(); source != nil {
		courier = &servicepb.DeliveryCourier{
			CourierId:     source.GetCourierId(),
			Name:          source.GetName(),
			Phone:         source.GetPhone(),
			TransportType: enumName(source.GetTransportType().String(), "TRANSPORT_TYPE_"),
			Status:        enumName(source.GetStatus().String(), "COURIER_STATUS_"),
			LastActiveAt:  source.GetLastActiveAt(),
		}
		if location := source.GetCurrentLocation(); location != nil {
			courier.CurrentLocation = &servicepb.DeliveryLocation{
				Latitude:  location.GetLatitude(),
				Longitude: location.GetLongitude(),
			}
		}
	}

	tracking := &servicepb.DeliveryTracking{
		OrderId:            resp.GetOrderId(),
		PackageId:          resp.GetPackageId(),
		Status:             enumName(resp.GetStatus().String(), "PACKAGE_STATUS_"),
		Courier:            courier,
		EstimatedArrivalAt: resp.GetEstimatedArrivalAt(),
		AssignedAt:         resp.GetAssignedAt(),
		DeliveredAt:        resp.GetDeliveredAt(),
	}

	if resp.EstimatedMinutesRemaining != nil {
		tracking.EstimatedMinutesRemaining = resp.EstimatedMinutesRemaining
	}
	if resp.DistanceKmRemaining != nil {
		tracking.DistanceKmRemaining = resp.DistanceKmRemaining
	}

	return tracking
}

func enumName(value, prefix string) string {
	return strings.TrimPrefix(value, prefix)
}

func grpcError(err error) error {
	if err == nil {
		return nil
	}
	st, ok := grpcstatus.FromError(err)
	if !ok {
		return err
	}
	switch st.Code() {
	case grpccodes.InvalidArgument:
		return connect.NewError(connect.CodeInvalidArgument, err)
	case grpccodes.NotFound:
		return connect.NewError(connect.CodeNotFound, err)
	case grpccodes.AlreadyExists:
		return connect.NewError(connect.CodeAlreadyExists, err)
	case grpccodes.PermissionDenied:
		return connect.NewError(connect.CodePermissionDenied, err)
	case grpccodes.FailedPrecondition:
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case grpccodes.Unavailable:
		return connect.NewError(connect.CodeUnavailable, err)
	case grpccodes.Unauthenticated:
		return connect.NewError(connect.CodeUnauthenticated, err)
	default:
		return connect.NewError(connect.CodeUnknown, err)
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
		return "", fmt.Errorf("%w", errEmptyGRPCTarget)
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
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
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
