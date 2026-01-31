//go:generate wire
//go:build wireinject

// The build tag makes sure the stub is not built in the final build.

/*
OMS DI-package
*/
package oms_di

import (
	"context"

	"github.com/authzed/authzed-go/v1"
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/client"

	"github.com/shortlink-org/go-sdk/auth/permission"
	config "github.com/shortlink-org/go-sdk/config"
	sdkctx "github.com/shortlink-org/go-sdk/context"
	"github.com/shortlink-org/go-sdk/db"
	"github.com/shortlink-org/go-sdk/flags"
	grpc "github.com/shortlink-org/go-sdk/grpc"
	logger "github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/go-sdk/observability/metrics"
	profiling "github.com/shortlink-org/go-sdk/observability/profiling"
	"github.com/shortlink-org/go-sdk/observability/tracing"
	"github.com/shortlink-org/go-sdk/temporal"

	"github.com/redis/rueidis"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/events"
	cartRepo "github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/cart"
	orderRepo "github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order"
	cartGoodsIndex "github.com/shortlink-org/shop/oms/internal/infrastructure/repository/redis/cart_goods_index"
	cartRPC "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1"
	orderRPC "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/run"
	temporalInfra "github.com/shortlink-org/shop/oms/internal/infrastructure/temporal"
	pguow "github.com/shortlink-org/shop/oms/pkg/uow/postgres"

	// Cart handlers
	cartAddItems "github.com/shortlink-org/shop/oms/internal/usecases/cart/command/add_items"
	cartRemoveItems "github.com/shortlink-org/shop/oms/internal/usecases/cart/command/remove_items"
	cartReset "github.com/shortlink-org/shop/oms/internal/usecases/cart/command/reset"
	cartGet "github.com/shortlink-org/shop/oms/internal/usecases/cart/query/get"

	// Order handlers
	orderCancel "github.com/shortlink-org/shop/oms/internal/usecases/order/command/cancel"
	orderCreate "github.com/shortlink-org/shop/oms/internal/usecases/order/command/create"
	orderUpdateDeliveryInfo "github.com/shortlink-org/shop/oms/internal/usecases/order/command/update_delivery_info"
	orderGet "github.com/shortlink-org/shop/oms/internal/usecases/order/query/get"

	// Checkout handlers
	checkout "github.com/shortlink-org/shop/oms/internal/usecases/checkout/command/create_order_from_cart"
)

type OMSService struct {
	// Common
	Log    logger.Logger
	Config *config.Config

	// Observability
	Tracer        trace.TracerProvider
	Monitoring    *metrics.Monitoring
	PprofEndpoint profiling.PprofEndpoint

	// Security
	authPermission *authzed.Client

	// Database
	DB db.DB

	// UnitOfWork
	UoW ports.UnitOfWork

	// Repositories
	CartRepo  ports.CartRepository
	OrderRepo ports.OrderRepository

	// Delivery
	run            *run.Response
	cartRPCServer  *cartRPC.CartRPC
	orderRPCServer *orderRPC.OrderRPC

	// Event Infrastructure
	EventPublisher *events.InMemoryPublisher

	// Temporal
	temporalClient client.Client
}

// OMSService ==========================================================================================================
// CustomDefaultSet - DefaultSet with go-sdk packages (config, context, flags, profiling)
var CustomDefaultSet = wire.NewSet(
	sdkctx.New,
	flags.New,
	newGoSDKProfiling,
	permission.New, // For authzed.Client
)

var OMSSet = wire.NewSet(
	// Common (custom DefaultSet)
	CustomDefaultSet,
	newGRPCServer,

	// Config & Observability (go-sdk)
	newGoSDKConfig,
	newGoSDKLogger,

	// Observability (go-sdk)
	newGoSDKTracer,
	newGoSDKMonitoring,

	// Database
	newDatabase,
	wire.FieldsOf(new(*metrics.Monitoring), "Metrics"),

	// Redis
	newRedisClient,

	// UnitOfWork
	newUnitOfWork,
	wire.Bind(new(ports.UnitOfWork), new(*pguow.UoW)),

	// Repositories
	newCartRepository,
	newOrderRepository,
	wire.Bind(new(ports.CartRepository), new(*cartRepo.Store)),
	wire.Bind(new(ports.OrderRepository), new(*orderRepo.Store)),

	// Indexes
	newCartGoodsIndex,
	wire.Bind(new(ports.CartGoodsIndex), new(*cartGoodsIndex.Store)),

	// Event Infrastructure
	newEventPublisher,
	wire.Bind(new(ports.EventPublisher), new(*events.InMemoryPublisher)),

	// Cart Handlers (concrete types, wire doesn't support generics)
	newCartAddItemsHandler,
	newCartRemoveItemsHandler,
	newCartResetHandler,
	newCartGetHandler,

	// Order Handlers (concrete types)
	newOrderCreateHandler,
	newOrderCancelHandler,
	newOrderUpdateDeliveryInfoHandler,
	newOrderGetHandler,

	// Checkout Handlers
	newCheckoutHandler,

	// Delivery
	newCartRPC,
	newOrderRPC,
	NewRunRPCServer,

	// Temporal
	temporal.New,
	newOrderEventSubscriber,

	NewOMSService,
)

// newGRPCServer creates a gRPC server using go-sdk/grpc
func newGRPCServer(ctx context.Context, log logger.Logger, tracer trace.TracerProvider, monitoring *metrics.Monitoring, cfg *config.Config) (*grpc.Server, error) {
	return grpc.InitServer(ctx, log, tracer, monitoring.Prometheus, nil, cfg)
}

// NewRunRPCServer starts the gRPC server
func NewRunRPCServer(runRPCServer *grpc.Server, _ *cartRPC.CartRPC) (*run.Response, error) {
	return run.Run(runRPCServer)
}

// newGoSDKConfig creates a go-sdk config instance
func newGoSDKConfig() (*config.Config, error) {
	return config.New()
}

// newGoSDKLogger creates a go-sdk logger instance for observability
func newGoSDKLogger(ctx context.Context, cfg *config.Config) (logger.Logger, func(), error) {
	return logger.NewDefault(ctx, cfg)
}

// newGoSDKTracer creates a tracer using go-sdk observability
func newGoSDKTracer(ctx context.Context, log logger.Logger, cfg *config.Config) (trace.TracerProvider, func(), error) {
	return tracing.New(ctx, log, cfg)
}

// newGoSDKMonitoring creates monitoring using go-sdk observability
func newGoSDKMonitoring(ctx context.Context, log logger.Logger, tracer trace.TracerProvider, cfg *config.Config) (*metrics.Monitoring, func(), error) {
	return metrics.New(ctx, log, tracer, cfg)
}

// newGoSDKProfiling creates profiling endpoint using go-sdk observability
func newGoSDKProfiling(ctx context.Context, log logger.Logger, tracer trace.TracerProvider, cfg *config.Config) (profiling.PprofEndpoint, error) {
	return profiling.New(ctx, log, tracer, cfg)
}

// newDatabase creates a database connection using go-sdk/db
func newDatabase(ctx context.Context, log logger.Logger, tracer trace.TracerProvider, meterProvider *sdkmetric.MeterProvider, cfg *config.Config) (db.DB, error) {
	return db.New(ctx, log, tracer, meterProvider, cfg)
}

// newUnitOfWork creates a PostgreSQL UnitOfWork
func newUnitOfWork(store db.DB) (*pguow.UoW, error) {
	pool, ok := store.GetConn().(*pgxpool.Pool)
	if !ok {
		return nil, db.ErrGetConnection
	}
	return pguow.New(pool), nil
}

// newCartRepository creates a PostgreSQL cart repository
func newCartRepository(ctx context.Context, store db.DB) (*cartRepo.Store, error) {
	return cartRepo.New(ctx, store)
}

// newOrderRepository creates a PostgreSQL order repository
func newOrderRepository(ctx context.Context, store db.DB) (*orderRepo.Store, error) {
	return orderRepo.New(ctx, store)
}

// newRedisClient creates a Redis client using rueidis
func newRedisClient(cfg *config.Config) (rueidis.Client, func(), error) {
	// Get Redis URI from config (default: localhost:6379)
	redisURI := cfg.GetString("STORE_REDIS_URI")
	if redisURI == "" {
		redisURI = "localhost:6379"
	}

	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{redisURI},
	})
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		client.Close()
	}

	return client, cleanup, nil
}

// newCartGoodsIndex creates a Redis-backed cart goods index
func newCartGoodsIndex(client rueidis.Client) *cartGoodsIndex.Store {
	return cartGoodsIndex.New(client)
}

// newEventPublisher creates an in-memory event publisher
func newEventPublisher() *events.InMemoryPublisher {
	return events.NewInMemoryPublisher()
}

// Cart Handler Factories
func newCartAddItemsHandler(log logger.Logger, uow ports.UnitOfWork, cartRepo ports.CartRepository, goodsIndex ports.CartGoodsIndex) *cartAddItems.Handler {
	return cartAddItems.NewHandler(log, uow, cartRepo, goodsIndex)
}

func newCartRemoveItemsHandler(log logger.Logger, uow ports.UnitOfWork, cartRepo ports.CartRepository, goodsIndex ports.CartGoodsIndex) *cartRemoveItems.Handler {
	return cartRemoveItems.NewHandler(log, uow, cartRepo, goodsIndex)
}

func newCartResetHandler(log logger.Logger, uow ports.UnitOfWork, cartRepo ports.CartRepository, goodsIndex ports.CartGoodsIndex) *cartReset.Handler {
	return cartReset.NewHandler(log, uow, cartRepo, goodsIndex)
}

func newCartGetHandler(uow ports.UnitOfWork, cartRepo ports.CartRepository) *cartGet.Handler {
	return cartGet.NewHandler(uow, cartRepo)
}

// Order Handler Factories
func newOrderCreateHandler(log logger.Logger, uow ports.UnitOfWork, orderRepo ports.OrderRepository, publisher ports.EventPublisher) *orderCreate.Handler {
	return orderCreate.NewHandler(log, uow, orderRepo, publisher)
}

func newOrderCancelHandler(log logger.Logger, uow ports.UnitOfWork, orderRepo ports.OrderRepository, publisher ports.EventPublisher) *orderCancel.Handler {
	return orderCancel.NewHandler(log, uow, orderRepo, publisher)
}

func newOrderUpdateDeliveryInfoHandler(log logger.Logger, uow ports.UnitOfWork, orderRepo ports.OrderRepository, publisher ports.EventPublisher) *orderUpdateDeliveryInfo.Handler {
	return orderUpdateDeliveryInfo.NewHandler(log, uow, orderRepo, publisher)
}

func newOrderGetHandler(uow ports.UnitOfWork, orderRepo ports.OrderRepository) *orderGet.Handler {
	return orderGet.NewHandler(uow, orderRepo)
}

// Checkout Handler Factory
func newCheckoutHandler(log logger.Logger, uow ports.UnitOfWork, cartRepo ports.CartRepository, orderRepo ports.OrderRepository, publisher ports.EventPublisher) *checkout.Handler {
	return checkout.NewHandler(log, uow, cartRepo, orderRepo, publisher)
}

// newOrderEventSubscriber creates and registers the order event subscriber
func newOrderEventSubscriber(log logger.Logger, temporalClient client.Client, publisher *events.InMemoryPublisher) *temporalInfra.OrderEventSubscriber {
	subscriber := temporalInfra.NewOrderEventSubscriber(log, temporalClient)
	subscriber.Register(publisher)
	return subscriber
}

// newCartRPC creates the Cart RPC server with handlers
func newCartRPC(
	runRPCServer *grpc.Server,
	log logger.Logger,
	addItemsHandler *cartAddItems.Handler,
	removeItemsHandler *cartRemoveItems.Handler,
	resetHandler *cartReset.Handler,
	getHandler *cartGet.Handler,
) (*cartRPC.CartRPC, error) {
	return cartRPC.New(runRPCServer, log, addItemsHandler, removeItemsHandler, resetHandler, getHandler)
}

// newOrderRPC creates the Order RPC server with handlers
func newOrderRPC(
	runRPCServer *grpc.Server,
	log logger.Logger,
	createHandler *orderCreate.Handler,
	cancelHandler *orderCancel.Handler,
	updateDeliveryInfoHandler *orderUpdateDeliveryInfo.Handler,
	checkoutHandler *checkout.Handler,
	getHandler *orderGet.Handler,
) (*orderRPC.OrderRPC, error) {
	return orderRPC.New(runRPCServer, log, createHandler, cancelHandler, updateDeliveryInfoHandler, checkoutHandler, getHandler)
}

func NewOMSService(
	// Common
	log logger.Logger,
	config *config.Config,

	// Observability
	monitoring *metrics.Monitoring,
	tracer trace.TracerProvider,
	pprofHTTP profiling.PprofEndpoint,

	// Security
	authPermission *authzed.Client,

	// Database
	database db.DB,

	// UnitOfWork
	unitOfWork ports.UnitOfWork,

	// Repositories
	cartRepository ports.CartRepository,
	orderRepository ports.OrderRepository,

	// Event Infrastructure
	eventPublisher *events.InMemoryPublisher,

	// Delivery
	run *run.Response,
	cartRPCServer *cartRPC.CartRPC,
	orderRPCServer *orderRPC.OrderRPC,

	// Temporal
	temporalClient client.Client,
	_ *temporalInfra.OrderEventSubscriber, // ensure subscriber is created
) (*OMSService, error) {
	return &OMSService{
		// Common
		Log:    log,
		Config: config,

		// Observability
		Tracer:        tracer,
		Monitoring:    monitoring,
		PprofEndpoint: pprofHTTP,

		// Security
		// TODO: enable later
		// authPermission: authPermission,

		// Database
		DB: database,

		// UnitOfWork
		UoW: unitOfWork,

		// Repositories
		CartRepo:  cartRepository,
		OrderRepo: orderRepository,

		// Event Infrastructure
		EventPublisher: eventPublisher,

		// Delivery
		run:            run,
		cartRPCServer:  cartRPCServer,
		orderRPCServer: orderRPCServer,

		// Temporal
		temporalClient: temporalClient,
	}, nil
}

func InitializeOMSService() (*OMSService, func(), error) {
	panic(wire.Build(OMSSet))
}
