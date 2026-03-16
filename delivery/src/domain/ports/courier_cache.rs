//! Courier Cache Port
//!
//! Defines the interface for caching courier state (hot data).
//! This port is implemented by infrastructure adapters (e.g., Redis).
//!
//! The cache stores frequently accessed/updated data:
//! - Current status (free/busy/unavailable)
//! - Current load and capacity
//! - Rating and delivery stats
//! - Sets of free couriers for quick lookup

use async_trait::async_trait;
#[cfg(test)]
use mockall::automock;
use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::courier::{Courier, CourierStatus};

/// Cached courier state (hot data that changes frequently)
#[derive(Debug, Clone)]
pub struct CachedCourierState {
    /// Current status
    pub status: CourierStatus,
    /// Current number of packages
    pub current_load: u32,
    /// Maximum number of packages
    pub max_load: u32,
    /// Performance rating (0.0 - 5.0)
    pub rating: f64,
    /// Number of successful deliveries
    pub successful_deliveries: u32,
    /// Number of failed deliveries
    pub failed_deliveries: u32,
}

impl From<&Courier> for CachedCourierState {
    fn from(courier: &Courier) -> Self {
        Self {
            status: courier.status(),
            current_load: courier.current_load(),
            max_load: courier.max_load(),
            rating: courier.rating(),
            successful_deliveries: courier.successful_deliveries(),
            failed_deliveries: courier.failed_deliveries(),
        }
    }
}

/// Errors that can occur during cache operations
#[derive(Debug, Error)]
pub enum CacheError {
    /// Connection error
    #[error("Cache connection error: {0}")]
    ConnectionError(String),

    /// Key not found
    #[error("Key not found: {0}")]
    NotFound(String),

    /// Serialization error
    #[error("Serialization error: {0}")]
    SerializationError(String),

    /// Operation error
    #[error("Cache operation error: {0}")]
    OperationError(String),
}

/// Courier Cache Port
///
/// Defines the contract for caching courier state.
/// Implementations handle the actual caching mechanism (Redis, etc.).
#[cfg_attr(test, automock)]
#[async_trait]
pub trait CourierCache: Send + Sync {
    /// Cache a full courier snapshot.
    ///
    /// Redis is a cache only; writes should originate from the aggregate and be
    /// mirrored here as a complete snapshot.
    async fn cache(&self, courier: &Courier) -> Result<(), CacheError>;

    /// Get courier state from cache
    async fn get_state(&self, courier_id: Uuid) -> Result<Option<CachedCourierState>, CacheError>;

    /// Get courier status
    async fn get_status(&self, courier_id: Uuid) -> Result<Option<CourierStatus>, CacheError>;

    /// Get all free courier IDs in a zone
    ///
    /// Returns IDs of couriers with status = Free in the specified zone.
    async fn get_free_couriers_in_zone(&self, zone: &str) -> Result<Vec<Uuid>, CacheError>;

    /// Get all free courier IDs (across all zones)
    async fn get_all_free_couriers(&self) -> Result<Vec<Uuid>, CacheError>;

    /// Remove courier from cache
    ///
    /// Called when a courier is deleted.
    async fn remove(&self, courier_id: Uuid, work_zone: &str) -> Result<(), CacheError>;

    /// Check if courier exists in cache
    async fn exists(&self, courier_id: Uuid) -> Result<bool, CacheError>;
}
