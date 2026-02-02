//go:generate wire
//go:build wireinject

// The build tag makes sure the stub is not built in the final build.

/*
OMS DI-package
*/
package oms_di

import (
	"github.com/authzed/authzed-go/v1"
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/client"

	"github.com/shortlink-org/go-sdk/auth/permission"
	config "github.com/shortlink-org/go-sdk/config"
	sdkctx "github.com/shortlink-org/go-sdk/context"
	"github.com/shortlink-org/go-sdk/db"
	"github.com/shortlink-org/go-sdk/flags"
	"github.com/shortlink-org/go-sdk/flight_trace"
	grpc "github.com/shortlink-org/go-sdk/grpc"
	logger "github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/go-sdk/observability/metrics"
	profiling "github.com/shortlink-org/go-sdk/observability/profiling"
	"github.com/shortlink-org/go-sdk/observability/tracing"
	"github.com/shortlink-org/go-sdk/temporal"

	"github.com/redis/rueidis"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/events"
	omsKafka "github.com/shortlink-org/shop/oms/internal/infrastructure/kafka"
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
	orderList "github.com/shortlink-org/shop/oms/internal/usecases/order/query/list"

	// Checkout handlers
	checkout "github.com/shortlink-org/shop/oms/internal/usecases/checkout/command/create_order_from_cart"

	// Temporal workers
	cart_worker "github.com/shortlink-org/shop/oms/internal/workers/cart/cart_worker"
	"github.com/shortlink-org/shop/oms/internal/workers/order/activities"
	order_worker "github.com/shortlink-org/shop/oms/internal/workers/order/order_worker"
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

	// Delivery Integration
	DeliveryClient   ports.DeliveryClient
	DeliveryConsumer *omsKafka.DeliveryConsumer

	// Pricer Integration
	PricerClient ports.PricerClient

	// Temporal
	temporalClient client.Client
	cartWorker     cart_worker.CartWorker
	orderWorker    order_worker.OrderWorker
}

// OMSService ==========================================================================================================
// CustomDefaultSet - DefaultSet with go-sdk packages (config, context, flags, profiling)
var CustomDefaultSet = wire.NewSet(
	sdkctx.New,
	flags.New,
	profiling.New,
	permission.New, // For authzed.Client
)

var OMSSet = wire.NewSet(
	// Common (custom DefaultSet)
	CustomDefaultSet,
	flight_trace.New,
	grpc.InitServer,

	// Config & Observability (go-sdk)
	config.New,
	logger.NewDefault,

	// Observability (go-sdk)
	tracing.New,
	metrics.New,

	// Database
	db.New,
	wire.FieldsOf(new(*metrics.Monitoring), "Metrics", "Prometheus"),

	// Redis
	newRedisClient,

	// UnitOfWork
	newUnitOfWork,
	wire.Bind(new(ports.UnitOfWork), new(*pguow.UoW)),

	// Repositories
	cartRepo.New,
	orderRepo.New,
	wire.Bind(new(ports.CartRepository), new(*cartRepo.Store)),
	wire.Bind(new(ports.OrderRepository), new(*orderRepo.Store)),

	// Indexes
	cartGoodsIndex.New,
	wire.Bind(new(ports.CartGoodsIndex), new(*cartGoodsIndex.Store)),

	// Event Infrastructure
	events.NewInMemoryPublisher,
	wire.Bind(new(ports.EventPublisher), new(*events.InMemoryPublisher)),

	// Delivery Integration (gRPC client + Kafka consumer)
	NewDeliveryClient,
	NewDeliveryConsumer,

	// Pricer Integration
	NewPricerClient,

	// Cart Handlers
	cartAddItems.NewHandler,
	cartRemoveItems.NewHandler,
	cartReset.NewHandler,
	cartGet.NewHandler,

	// Order Handlers
	orderCreate.NewHandler,
	orderCancel.NewHandler,
	orderUpdateDeliveryInfo.NewHandler,
	orderGet.NewHandler,
	orderList.NewHandler,

	// Checkout Handlers
	checkout.NewHandler,

	// Delivery
	cartRPC.New,
	orderRPC.New,
	NewRunRPCServer,

	// Temporal
	temporal.New,
	newOrderEventSubscriber,

	// Temporal Workers
	cart_worker.New,
	activities.New,
	order_worker.NewWithActivities,

	NewOMSService,
)

// NewRunRPCServer starts the gRPC server
func NewRunRPCServer(runRPCServer *grpc.Server, _ *cartRPC.CartRPC) (*run.Response, error) {
	return run.Run(runRPCServer)
}

// newUnitOfWork creates a PostgreSQL UnitOfWork
func newUnitOfWork(store db.DB) (*pguow.UoW, error) {
	pool, ok := store.GetConn().(*pgxpool.Pool)
	if !ok {
		return nil, db.ErrGetConnection
	}
	return pguow.New(pool), nil
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

// newOrderEventSubscriber creates and registers the order event subscriber
func newOrderEventSubscriber(log logger.Logger, temporalClient client.Client, publisher *events.InMemoryPublisher) *temporalInfra.OrderEventSubscriber {
	subscriber := temporalInfra.NewOrderEventSubscriber(log, temporalClient)
	subscriber.Register(publisher)
	return subscriber
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

	// Delivery Integration
	deliveryClient ports.DeliveryClient,
	deliveryConsumer *omsKafka.DeliveryConsumer,

	// Pricer Integration
	pricerClient ports.PricerClient,

	// gRPC Servers
	run *run.Response,
	cartRPCServer *cartRPC.CartRPC,
	orderRPCServer *orderRPC.OrderRPC,

	// Temporal
	temporalClient client.Client,
	_ *temporalInfra.OrderEventSubscriber, // ensure subscriber is created
	cartWorker cart_worker.CartWorker,
	orderWorker order_worker.OrderWorker,
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

		// Delivery Integration
		DeliveryClient:   deliveryClient,
		DeliveryConsumer: deliveryConsumer,

		// Pricer Integration
		PricerClient: pricerClient,

		// gRPC Servers
		run:            run,
		cartRPCServer:  cartRPCServer,
		orderRPCServer: orderRPCServer,

		// Temporal
		temporalClient: temporalClient,
		cartWorker:     cartWorker,
		orderWorker:    orderWorker,
	}, nil
}

func InitializeOMSService() (*OMSService, func(), error) {
	panic(wire.Build(OMSSet))
}
