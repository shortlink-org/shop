//! Dependency Injection Module
//!
//! Provides application state and dependency wiring.

use std::sync::Arc;

use redis::aio::ConnectionManager;
use sea_orm::{Database, DatabaseConnection};
use thiserror::Error;
use tracing::info;

use crate::config::Config;
use crate::infrastructure::cache::CourierRedisCache;
use crate::infrastructure::repository::{CourierPostgresRepository, PackagePostgresRepository};

/// DI initialization errors
#[derive(Debug, Error)]
pub enum DiError {
    #[error("Database connection failed: {0}")]
    DatabaseError(String),

    #[error("Redis connection failed: {0}")]
    RedisError(String),
}

/// Application state containing all dependencies
pub struct AppState {
    /// PostgreSQL courier repository
    pub courier_repo: Arc<CourierPostgresRepository>,

    /// PostgreSQL package repository
    pub package_repo: Arc<PackagePostgresRepository>,

    /// Redis courier cache
    pub courier_cache: Arc<CourierRedisCache>,

    /// Database connection (for migrations, etc.)
    pub db: DatabaseConnection,
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

        // Create repositories and caches
        let courier_repo = Arc::new(CourierPostgresRepository::new(db.clone()));
        let package_repo = Arc::new(PackagePostgresRepository::new(db.clone()));
        let courier_cache = Arc::new(CourierRedisCache::new(redis_conn));

        info!("Application state initialized");

        Ok(Self {
            courier_repo,
            package_repo,
            courier_cache,
            db,
        })
    }
}
