//! Location Cache Port
//!
//! Interface for caching location data in Redis.

use async_trait::async_trait;
use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::CourierLocation;

/// Cache errors
#[derive(Debug, Error)]
pub enum CacheError {
    #[error("Connection error: {0}")]
    ConnectionError(String),

    #[error("Serialization error: {0}")]
    SerializationError(String),

    #[error("Operation error: {0}")]
    OperationError(String),
}

/// Location Cache trait
#[async_trait]
pub trait LocationCache: Send + Sync {
    /// Set courier's current location in cache
    /// TTL is typically 5 minutes
    async fn set_location(&self, location: &CourierLocation) -> Result<(), CacheError>;

    /// Get courier's current location from cache
    async fn get_location(&self, courier_id: Uuid) -> Result<Option<CourierLocation>, CacheError>;

    /// Get multiple couriers' locations from cache (batch get)
    async fn get_locations(&self, courier_ids: &[Uuid]) -> Result<Vec<(Uuid, Option<CourierLocation>)>, CacheError>;

    /// Remove courier's location from cache
    async fn remove_location(&self, courier_id: Uuid) -> Result<(), CacheError>;

    /// Check if courier's location exists in cache
    async fn exists(&self, courier_id: Uuid) -> Result<bool, CacheError>;
}
