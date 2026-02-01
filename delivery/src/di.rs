//! Dependency Injection Module
//!
//! Provides application state and dependency wiring.

use std::sync::Arc;

use redis::aio::ConnectionManager;
use sea_orm::{Database, DatabaseConnection};
use thiserror::Error;
use tokio::sync::broadcast;
use tracing::info;

use crate::config::Config;
use crate::infrastructure::cache::{CourierRedisCache, RedisLocationCache};
use crate::infrastructure::messaging::{
    kafka_publisher::KafkaPublisherConfig, KafkaEventPublisher, LocationConsumer,
    location_consumer::LocationConsumerConfig,
};
use crate::infrastructure::notifications::StubNotificationService;
use crate::infrastructure::repository::{
    CourierPostgresRepository, LocationPostgresRepository, PackagePostgresRepository,
};

/// DI initialization errors
#[derive(Debug, Error)]
pub enum DiError {
    #[error("Database connection failed: {0}")]
    DatabaseError(String),

    #[error("Redis connection failed: {0}")]
    RedisError(String),

    #[error("Kafka connection failed: {0}")]
    KafkaError(String),
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

    /// Kafka event publisher
    pub event_publisher: Arc<KafkaEventPublisher>,

    /// Notification service (stub for now)
    pub notification_service: Arc<StubNotificationService>,

    /// Database connection (for migrations, etc.)
    pub db: DatabaseConnection,

    /// Shutdown signal sender
    pub shutdown_tx: broadcast::Sender<()>,
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

        // Connect to Redis
        info!("Connecting to Redis...");
        let redis_client = redis::Client::open(config.redis_url.as_str())
            .map_err(|e| DiError::RedisError(e.to_string()))?;
        let redis_conn = ConnectionManager::new(redis_client)
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
        let courier_cache = Arc::new(CourierRedisCache::new(redis_conn.clone()));
        let location_cache = Arc::new(RedisLocationCache::new(redis_conn));

        // Create notification service (stub for now)
        let notification_service = Arc::new(StubNotificationService::new());

        // Create shutdown channel
        let (shutdown_tx, _) = broadcast::channel(1);

        info!("Application state initialized");

        Ok(Self {
            courier_repo,
            package_repo,
            location_repo,
            courier_cache,
            location_cache,
            event_publisher,
            notification_service,
            db,
            shutdown_tx,
        })
    }

    /// Start background consumers
    pub async fn start_consumers(&self) -> Result<(), DiError> {
        info!("Starting location consumer...");

        let consumer_config = LocationConsumerConfig::from_env();
        let location_cache = self.location_cache.clone();
        let location_repo = self.location_repo.clone();
        let shutdown_rx = self.shutdown_tx.subscribe();

        // Create and start location consumer
        let consumer = LocationConsumer::new(
            consumer_config,
            location_cache,
            location_repo,
            shutdown_rx,
        )
        .map_err(|e| DiError::KafkaError(e))?;

        // Spawn consumer as background task
        tokio::spawn(async move {
            consumer.run().await;
        });

        info!("Location consumer started");

        Ok(())
    }

    /// Shutdown the application
    pub fn shutdown(&self) {
        info!("Sending shutdown signal...");
        let _ = self.shutdown_tx.send(());
    }
}
