//! Location Cache Port
//!
//! Defines the interface for caching current courier locations (hot data).
//! This port is implemented by infrastructure adapters (e.g., Redis).
//!
//! The cache stores:
//! - Current courier position (refreshed every few seconds)
//! - With TTL to auto-expire stale locations

use async_trait::async_trait;
#[cfg(test)]
use mockall::automock;
use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::CourierLocation;

/// Errors that can occur during location cache operations
#[derive(Debug, Error)]
pub enum LocationCacheError {
    /// Connection error
    #[error("Cache connection error: {0}")]
    ConnectionError(String),

    /// Location not found in cache
    #[error("Location not found for courier: {0}")]
    NotFound(Uuid),

    /// Serialization error
    #[error("Serialization error: {0}")]
    SerializationError(String),

    /// Operation error
    #[error("Cache operation error: {0}")]
    OperationError(String),
}

/// Location Cache Port
///
/// Defines the contract for caching courier current locations.
/// Implementations handle the actual caching mechanism (Redis, etc.).
#[cfg_attr(test, automock)]
#[async_trait]
pub trait LocationCache: Send + Sync {
    /// Set courier's current location
    ///
    /// Stores the location with a TTL (e.g., 5 minutes).
    /// If the courier doesn't report a new location within the TTL,
    /// the cached location will expire automatically.
    async fn set_location(
        &self,
        location: &CourierLocation,
    ) -> Result<(), LocationCacheError>;

    /// Get courier's current location
    ///
    /// Returns the cached location or None if not found/expired.
    async fn get_location(
        &self,
        courier_id: Uuid,
    ) -> Result<Option<CourierLocation>, LocationCacheError>;

    /// Get locations for multiple couriers
    ///
    /// Returns only the locations that exist in cache.
    async fn get_locations(
        &self,
        courier_ids: &[Uuid],
    ) -> Result<Vec<CourierLocation>, LocationCacheError>;

    /// Delete courier's location from cache
    async fn delete_location(&self, courier_id: Uuid) -> Result<(), LocationCacheError>;

    /// Check if courier's location exists in cache
    async fn has_location(&self, courier_id: Uuid) -> Result<bool, LocationCacheError>;

    /// Get all cached courier locations
    ///
    /// Returns all currently cached locations (for admin/monitoring).
    /// May be expensive - use with caution.
    async fn get_all_locations(&self) -> Result<Vec<CourierLocation>, LocationCacheError>;

    /// Get courier IDs with active (non-expired) locations
    ///
    /// Useful for finding all currently active couriers.
    async fn get_active_courier_ids(&self) -> Result<Vec<Uuid>, LocationCacheError>;
}
