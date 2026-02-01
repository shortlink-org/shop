//! Location Repository Port
//!
//! Interface for persisting and retrieving location data.

use async_trait::async_trait;
use chrono::{DateTime, Utc};
use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::{CourierLocation, Geofence, LocationHistoryEntry, TimeRange};

/// Repository errors
#[derive(Debug, Error)]
pub enum RepositoryError {
    #[error("Entity not found: {0}")]
    NotFound(String),

    #[error("Database error: {0}")]
    DatabaseError(String),

    #[error("Serialization error: {0}")]
    SerializationError(String),

    #[error("Query error: {0}")]
    QueryError(String),
}

/// Location Repository trait
#[async_trait]
pub trait LocationRepository: Send + Sync {
    // === Current Location ===

    /// Save or update courier's current location
    async fn save_current_location(&self, location: &CourierLocation) -> Result<(), RepositoryError>;

    /// Get courier's current location
    async fn get_current_location(&self, courier_id: Uuid) -> Result<Option<CourierLocation>, RepositoryError>;

    /// Get current locations for multiple couriers
    async fn get_current_locations(&self, courier_ids: &[Uuid]) -> Result<Vec<CourierLocation>, RepositoryError>;

    // === Location History ===

    /// Save a location history entry
    async fn save_history_entry(&self, entry: &LocationHistoryEntry) -> Result<(), RepositoryError>;

    /// Get location history for a courier within a time range
    async fn get_location_history(
        &self,
        courier_id: Uuid,
        time_range: &TimeRange,
        limit: Option<u32>,
    ) -> Result<Vec<LocationHistoryEntry>, RepositoryError>;

    /// Delete old history entries (for cleanup)
    async fn delete_history_before(&self, before: DateTime<Utc>) -> Result<u64, RepositoryError>;
}

/// Geofence Repository trait
#[async_trait]
pub trait GeofenceRepository: Send + Sync {
    /// Save a geofence
    async fn save(&self, geofence: &Geofence) -> Result<(), RepositoryError>;

    /// Get a geofence by ID
    async fn find_by_id(&self, id: Uuid) -> Result<Option<Geofence>, RepositoryError>;

    /// Get all active geofences
    async fn find_active(&self) -> Result<Vec<Geofence>, RepositoryError>;

    /// Get all geofences
    async fn find_all(&self, limit: u64, offset: u64) -> Result<Vec<Geofence>, RepositoryError>;

    /// Delete a geofence
    async fn delete(&self, id: Uuid) -> Result<(), RepositoryError>;
}
