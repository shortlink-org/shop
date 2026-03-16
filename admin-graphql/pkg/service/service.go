package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	servicepb "github.com/shortlink-org/shop/admin-graphql/pkg/generated/service/v1"
	serviceconnect "github.com/shortlink-org/shop/admin-graphql/pkg/generated/service/v1/v1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	defaultListenAddr  = "0.0.0.0:4012"
	defaultAdminAPI    = "http://127.0.0.1:8000"
	authHeader         = "Authorization"
	cookieHeader       = "Cookie"
	userIDHeader       = "X-User-ID"
	traceparentHeader  = "traceparent"
	traceIDHeader      = "trace-id"
)

type Service struct {
	logger      *slog.Logger
	adminAPIURL string
	httpClient  *http.Client
}

type goodDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Price       string `json:"price"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type goodsListDTO struct {
	Count    int32     `json:"count"`
	Next     *string   `json:"next"`
	Previous *string   `json:"previous"`
	Results  []goodDTO `json:"results"`
}

func New(logger *slog.Logger, adminAPIURL string, httpClient *http.Client) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout:   10 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		}
	} else if httpClient.Transport == nil {
		httpClient.Transport = otelhttp.NewTransport(http.DefaultTransport)
	}

	return &Service{
		logger:      logger,
		adminAPIURL: strings.TrimRight(adminAPIURL, "/"),
		httpClient:  httpClient,
	}
}

func Start(ctx context.Context) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	listenAddr := envOrDefault("LISTEN_ADDR", defaultListenAddr)
	adminAPIURL := envOrDefault("ADMIN_API_URL", defaultAdminAPI)
	shutdownTracing, err := setupTracing(ctx, "shortlink-shop-admin-graphql")
	if err != nil {
		return fmt.Errorf("setup tracing: %w", err)
	}
	defer func() {
		if shutdownErr := shutdownTracing(context.Background()); shutdownErr != nil {
			logger.Error("shutdown tracing", "error", shutdownErr)
		}
	}()
	svc := New(logger, adminAPIURL, nil)

	mux := http.NewServeMux()
	path, handler := serviceconnect.NewAdminHandler(svc)
	mux.Handle(path, handler)
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    listenAddr,
		Handler: otelhttp.NewHandler(h2c.NewHandler(mux, &http2.Server{}), "admin-graphql"),
	}

	logger.Info("starting admin-graphql grpc subgraph", "listen_addr", listenAddr, "admin_api_url", adminAPIURL)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown admin-graphql grpc subgraph", "error", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Service) QueryGoods(
	ctx context.Context,
	req *connect.Request[servicepb.QueryGoodsRequest],
) (*connect.Response[servicepb.QueryGoodsResponse], error) {
	var payload goodsListDTO
	if err := s.getJSON(ctx, req.Header(), "/goods/", &payload); err != nil {
		return nil, err
	}

	return connect.NewResponse(&servicepb.QueryGoodsResponse{
		Goods: mapGoodsList(payload),
	}), nil
}

func (s *Service) QueryGood(
	ctx context.Context,
	req *connect.Request[servicepb.QueryGoodRequest],
) (*connect.Response[servicepb.QueryGoodResponse], error) {
	if strings.TrimSpace(req.Msg.GetId()) == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id is required"))
	}

	var payload goodDTO
	if err := s.getJSON(ctx, req.Header(), "/goods/"+req.Msg.GetId()+"/", &payload); err != nil {
		return nil, err
	}

	return connect.NewResponse(&servicepb.QueryGoodResponse{
		Good: mapGood(payload),
	}), nil
}

func (s *Service) getJSON(ctx context.Context, headers http.Header, path string, out any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.adminAPIURL+path, nil)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}

	propagateHeader(headers, request.Header, authHeader)
	propagateHeader(headers, request.Header, cookieHeader)
	propagateHeader(headers, request.Header, userIDHeader)
	propagateHeader(headers, request.Header, traceparentHeader)
	propagateHeader(headers, request.Header, traceIDHeader)

	response, err := s.httpClient.Do(request)
	if err != nil {
		return connect.NewError(connect.CodeUnavailable, err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return connect.NewError(httpStatusToConnectCode(response.StatusCode), fmt.Errorf("admin api %s returned %d: %s", path, response.StatusCode, strings.TrimSpace(string(body))))
	}

	if err := json.NewDecoder(response.Body).Decode(out); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}

	return nil
}

func propagateHeader(from http.Header, to http.Header, name string) {
	if value := strings.TrimSpace(from.Get(name)); value != "" {
		to.Set(name, value)
	}
}

func httpStatusToConnectCode(status int) connect.Code {
	switch status {
	case http.StatusBadRequest:
		return connect.CodeInvalidArgument
	case http.StatusUnauthorized:
		return connect.CodeUnauthenticated
	case http.StatusForbidden:
		return connect.CodePermissionDenied
	case http.StatusNotFound:
		return connect.CodeNotFound
	case http.StatusConflict:
		return connect.CodeAlreadyExists
	case http.StatusTooManyRequests:
		return connect.CodeResourceExhausted
	case http.StatusNotImplemented:
		return connect.CodeUnimplemented
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return connect.CodeUnavailable
	default:
		if status >= 500 {
			return connect.CodeInternal
		}
		return connect.CodeUnknown
	}
}

func mapGoodsList(payload goodsListDTO) *servicepb.GoodsList {
	results := make([]*servicepb.Good, 0, len(payload.Results))
	for _, good := range payload.Results {
		results = append(results, mapGood(good))
	}

	var next *wrapperspb.StringValue
	if payload.Next != nil && *payload.Next != "" {
		next = wrapperspb.String(*payload.Next)
	}

	var previous *wrapperspb.StringValue
	if payload.Previous != nil && *payload.Previous != "" {
		previous = wrapperspb.String(*payload.Previous)
	}

	return &servicepb.GoodsList{
		Count:    payload.Count,
		Next:     next,
		Previous: previous,
		Results:  results,
	}
}

func mapGood(payload goodDTO) *servicepb.Good {
	return &servicepb.Good{
		Id:          payload.ID,
		Name:        payload.Name,
		Price:       payload.Price,
		Description: payload.Description,
		CreatedAt:   payload.CreatedAt,
		UpdatedAt:   payload.UpdatedAt,
	}
}

func envOrDefault(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
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
