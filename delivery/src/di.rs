//! Dependency Injection Module
//!
//! Provides application state and dependency wiring.

use std::collections::HashSet;
use std::sync::Arc;

use migration::{Migrator, MigratorTrait};
use redis::aio::ConnectionManager;
use sea_orm::{Database, DatabaseConnection};
use thiserror::Error;
use tokio::sync::broadcast;
use tracing::{info, warn};

use crate::config::{Config, RandomAddressBbox, TemporalConfig};
use crate::domain::model::package::{PackageId, PackageStatus};
use crate::domain::ports::PackageRepository;
use crate::infrastructure::cache::{CourierRedisCache, RedisLocationCache};
use crate::infrastructure::geolocation::StubGeolocationService;
use crate::infrastructure::messaging::{
    emulation_consumer::EmulationConsumerConfig, kafka_publisher::KafkaPublisherConfig,
    location_consumer::LocationConsumerConfig, outbox_forwarder::OutboxForwarderConfig,
    EmulationConsumer, KafkaEventPublisher, LocationConsumer, OutboxForwarder, RedisTrackingPubSub,
};
use crate::infrastructure::notifications::StubNotificationService;
use crate::infrastructure::osrm::OsrmClient;
use crate::infrastructure::repository::{
    CourierPostgresRepository, LocationPostgresRepository, OutboxPostgresRepository,
    PackagePostgresRepository,
};
use crate::usecases::courier::command::{
    AcceptPackageHandler, CompleteCourierDeliveryHandler, RegisterHandler,
};
use crate::usecases::courier::query::get_pool::Handler as GetPoolHandler;
use crate::workers::courier::CourierActivities;
use crate::workers::delivery::DeliveryActivities;
use crate::workers::TemporalWorkerRunner;

/// DI initialization errors
#[derive(Debug, Error)]
pub enum DiError {
    #[error("Database connection failed: {0}")]
    DatabaseError(String),

    #[error("Redis connection failed: {0}")]
    RedisError(String),

    #[error("Kafka connection failed: {0}")]
    KafkaError(String),

    #[error("OSRM client initialization failed: {0}")]
    OsrmError(String),

    #[error("Temporal worker error: {0}")]
    TemporalError(String),
}

/// Application state containing all dependencies
pub struct AppState {
    /// PostgreSQL courier repository
    pub courier_repo: Arc<CourierPostgresRepository>,

    /// PostgreSQL package repository
    pub package_repo: Arc<PackagePostgresRepository>,

    /// PostgreSQL location repository
    pub location_repo: Arc<LocationPostgresRepository>,

    /// Redis courier cache
    pub courier_cache: Arc<CourierRedisCache>,

    /// Redis location cache
    pub location_cache: Arc<RedisLocationCache>,

    /// Geolocation service (stub delegating to cache + repository)
    pub geolocation_service:
        Arc<StubGeolocationService<RedisLocationCache, LocationPostgresRepository>>,

    /// Kafka event publisher
    pub event_publisher: Arc<KafkaEventPublisher>,

    /// Notification service (stub for now)
    pub notification_service: Arc<StubNotificationService>,

    /// Transactional outbox repository
    pub outbox_repo: Arc<OutboxPostgresRepository>,

    /// Database connection (for migrations, etc.)
    pub db: DatabaseConnection,

    /// Shutdown signal sender
    pub shutdown_tx: broadcast::Sender<()>,

    /// Redis-backed tracking pub/sub for cross-pod subscription fan-out
    pub tracking_pubsub: Arc<RedisTrackingPubSub>,

    /// OSRM HTTP client for routing and nearest-road lookups
    pub osrm_client: Arc<OsrmClient>,

    /// Bounding box for GetRandomAddress (optional)
    pub random_address_bbox: Option<RandomAddressBbox>,
}

impl AppState {
    /// Create a new AppState with all dependencies initialized
    pub async fn new(config: &Config) -> Result<Self, DiError> {
        info!("Initializing application state...");

        // Connect to PostgreSQL
        info!("Connecting to PostgreSQL...");
        let db = Database::connect(&config.database_url)
            .await
            .map_err(|e| DiError::DatabaseError(e.to_string()))?;
        info!("PostgreSQL connected");

        // Run migrations
        info!("Running database migrations...");
        Migrator::up(&db, None)
            .await
            .map_err(|e| DiError::DatabaseError(format!("Migration failed: {}", e)))?;
        info!("Migrations completed");

        // Connect to Redis
        info!("Connecting to Redis...");
        let redis_client = redis::Client::open(config.redis_url.as_str())
            .map_err(|e| DiError::RedisError(e.to_string()))?;
        let redis_conn = ConnectionManager::new(redis_client.clone())
            .await
            .map_err(|e| DiError::RedisError(e.to_string()))?;
        info!("Redis connected");

        // Create Kafka event publisher
        info!("Connecting to Kafka...");
        let kafka_config = KafkaPublisherConfig::from_env();
        let event_publisher = Arc::new(
            KafkaEventPublisher::new(kafka_config)
                .map_err(|e| DiError::KafkaError(e.to_string()))?,
        );
        info!("Kafka event publisher connected");

        // Create repositories and caches
        let courier_repo = Arc::new(CourierPostgresRepository::new(db.clone()));
        let package_repo = Arc::new(PackagePostgresRepository::new(db.clone()));
        let location_repo = Arc::new(LocationPostgresRepository::new(db.clone()));
        let outbox_repo = Arc::new(OutboxPostgresRepository::new(db.clone()));
        let courier_cache = Arc::new(CourierRedisCache::new(redis_conn.clone()));
        let location_cache = Arc::new(RedisLocationCache::new(redis_conn.clone()));
        let geolocation_service = Arc::new(StubGeolocationService::new(
            location_cache.clone(),
            location_repo.clone(),
        ));
        let tracking_pubsub = Arc::new(RedisTrackingPubSub::new(redis_client, redis_conn));

        // Create notification service (stub for now)
        let notification_service = Arc::new(StubNotificationService::new());
        let osrm_client = Arc::new(
            OsrmClient::new(
                config.osrm.base_url.clone(),
                config.osrm.timeout,
                config.osrm.auth_header_name.as_deref(),
                config.osrm.auth_header_value.as_deref(),
            )
            .map_err(|e| DiError::OsrmError(e.to_string()))?,
        );

        // Create shutdown channel
        let (shutdown_tx, _) = broadcast::channel(1);

        info!("Application state initialized");

        Ok(Self {
            courier_repo,
            package_repo,
            location_repo,
            courier_cache,
            location_cache,
            geolocation_service,
            event_publisher,
            notification_service,
            outbox_repo,
            db,
            shutdown_tx,
            tracking_pubsub,
            osrm_client,
            random_address_bbox: config.random_address_bbox.clone(),
        })
    }

    /// Start background consumers
    pub async fn start_consumers(&self) -> Result<(), DiError> {
        info!("Starting delivery outbox forwarder...");

        let outbox_forwarder = OutboxForwarder::new(
            OutboxForwarderConfig::from_env(),
            self.outbox_repo.clone(),
            self.event_publisher.clone(),
            self.shutdown_tx.subscribe(),
        );

        tokio::spawn(async move {
            outbox_forwarder.run().await;
        });

        info!("Delivery outbox forwarder started");

        let mut kafka_errors = Vec::new();

        info!("Starting location consumer...");

        let consumer_config = LocationConsumerConfig::from_env();
        let location_cache = self.location_cache.clone();
        let location_repo = self.location_repo.clone();
        let package_repo = self.package_repo.clone();
        let tracking_pubsub = self.tracking_pubsub.clone();
        let shutdown_rx = self.shutdown_tx.subscribe();

        match LocationConsumer::new(
            consumer_config,
            location_cache,
            location_repo,
            package_repo,
            tracking_pubsub,
            shutdown_rx,
        ) {
            Ok(consumer) => {
                tokio::spawn(async move {
                    consumer.run().await;
                });
                info!("Location consumer started");
            }
            Err(err) => kafka_errors.push(format!("location consumer: {err}")),
        }

        info!("Starting courier-emulation consumer...");

        let emulation_consumer_config = EmulationConsumerConfig::from_env();
        let emulation_shutdown_rx = self.shutdown_tx.subscribe();

        match EmulationConsumer::new(
            emulation_consumer_config,
            self.courier_repo.clone(),
            self.courier_cache.clone(),
            self.package_repo.clone(),
            self.geolocation_service.clone(),
            self.tracking_pubsub.clone(),
            emulation_shutdown_rx,
        ) {
            Ok(emulation_consumer) => {
                tokio::spawn(async move {
                    emulation_consumer.run().await;
                });
                info!("Courier-emulation consumer started");
            }
            Err(err) => kafka_errors.push(format!("courier-emulation consumer: {err}")),
        }

        if kafka_errors.is_empty() {
            Ok(())
        } else {
            Err(DiError::KafkaError(kafka_errors.join("; ")))
        }
    }

    /// Shutdown the application
    pub fn shutdown(&self) {
        info!("Sending shutdown signal...");
        let _ = self.shutdown_tx.send(());
    }

    pub async fn publish_tracking_update(&self, order_id: uuid::Uuid) {
        if let Err(err) = self.tracking_pubsub.publish_order_update(order_id).await {
            warn!(order_id = %order_id, error = %err, "failed to publish tracking update");
        }
    }

    pub async fn publish_tracking_update_for_package(&self, package_id: uuid::Uuid) {
        let package_id = PackageId::from_uuid(package_id);
        if let Ok(Some(package)) = self.package_repo.find_by_id(package_id).await {
            self.publish_tracking_update(package.order_id()).await;
        }
    }

    pub async fn publish_tracking_updates_for_courier(&self, courier_id: uuid::Uuid) {
        let Ok(packages) = self.package_repo.find_by_courier(courier_id).await else {
            return;
        };

        let mut order_ids = HashSet::new();
        for package in packages {
            if matches!(
                package.status(),
                PackageStatus::Assigned | PackageStatus::InTransit
            ) {
                order_ids.insert(package.order_id());
            }
        }

        for order_id in order_ids {
            self.publish_tracking_update(order_id).await;
        }
    }

    /// Initialize Temporal worker configuration
    ///
    /// Validates that the Temporal runtime can be created.
    /// Note: Due to SDK limitations, workers should be run in dedicated processes.
    /// Use `run_courier_worker` or `run_delivery_worker` methods directly.
    pub async fn start_temporal_workers(
        self: &Arc<Self>,
        config: &TemporalConfig,
    ) -> Result<(), DiError> {
        info!("Initializing Temporal worker configuration...");

        // Validate that we can create a worker runner
        let _runner = TemporalWorkerRunner::new(config.clone())
            .map_err(|e| DiError::TemporalError(e.to_string()))?;

        info!(
            host = %config.host,
            namespace = %config.namespace,
            courier_queue = %config.task_queue_courier,
            delivery_queue = %config.task_queue_delivery,
            "Temporal configuration validated"
        );

        // Note: The Temporal Rust SDK pre-alpha has limitations with tokio::spawn
        // due to RefCell in Worker. Workers should be run in dedicated processes
        // or using tokio::task::spawn_local on a LocalSet.
        //
        // For production, consider:
        // 1. Running workers as separate binaries
        // 2. Using tokio::task::LocalSet for worker execution
        // 3. Waiting for SDK stabilization

        info!("Temporal workers ready (run in dedicated process for production)");

        Ok(())
    }

    /// Create courier activities for use with Temporal worker
    pub fn create_courier_activities(
        self: &Arc<Self>,
    ) -> Arc<CourierActivities<CourierPostgresRepository, CourierRedisCache>> {
        let register_handler = Arc::new(RegisterHandler::new(
            self.courier_repo.clone(),
            self.courier_cache.clone(),
        ));
        let get_pool_handler = Arc::new(GetPoolHandler::new(
            self.courier_repo.clone(),
            self.courier_cache.clone(),
        ));
        let accept_package_handler = Arc::new(AcceptPackageHandler::new(
            self.courier_repo.clone(),
            self.courier_cache.clone(),
        ));
        let complete_delivery_handler = Arc::new(CompleteCourierDeliveryHandler::new(
            self.courier_repo.clone(),
            self.courier_cache.clone(),
        ));

        Arc::new(CourierActivities::new(
            register_handler,
            get_pool_handler,
            accept_package_handler,
            complete_delivery_handler,
            self.courier_repo.clone(),
            self.courier_cache.clone(),
        ))
    }

    /// Create delivery activities for use with Temporal worker
    pub fn create_delivery_activities(
        self: &Arc<Self>,
    ) -> Arc<DeliveryActivities<CourierPostgresRepository, CourierRedisCache>> {
        let get_pool_handler = Arc::new(GetPoolHandler::new(
            self.courier_repo.clone(),
            self.courier_cache.clone(),
        ));
        let accept_package_handler = Arc::new(AcceptPackageHandler::new(
            self.courier_repo.clone(),
            self.courier_cache.clone(),
        ));
        let complete_delivery_handler = Arc::new(CompleteCourierDeliveryHandler::new(
            self.courier_repo.clone(),
            self.courier_cache.clone(),
        ));

        Arc::new(DeliveryActivities::new(
            get_pool_handler,
            accept_package_handler,
            complete_delivery_handler,
        ))
    }
}
