//! Location Repository Port
//!
//! Defines the interface for persisting courier location history.
//! This is the secondary port (driven) for location persistence.

use async_trait::async_trait;
#[cfg(test)]
use mockall::automock;
use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::courier_location::{LocationHistoryEntry, TimeRange};

/// Errors that can occur during location repository operations
#[derive(Debug, Error)]
pub enum LocationRepositoryError {
    /// Database connection error
    #[error("Database connection error: {0}")]
    ConnectionError(String),

    /// Record not found
    #[error("Location record not found")]
    NotFound,

    /// Database query error
    #[error("Database query error: {0}")]
    QueryError(String),

    /// Data integrity error
    #[error("Data integrity error: {0}")]
    DataError(String),
}

/// Location Repository Port
///
/// Defines the contract for storing and retrieving courier location history.
/// Implementations handle the actual database operations (PostgreSQL, etc.).
#[cfg_attr(test, automock)]
#[async_trait]
pub trait LocationRepository: Send + Sync {
    /// Save a location history entry
    ///
    /// Persists the location entry to the database.
    async fn save(&self, entry: &LocationHistoryEntry) -> Result<(), LocationRepositoryError>;

    /// Save multiple location entries in a batch
    ///
    /// More efficient for bulk inserts.
    async fn save_batch(
        &self,
        entries: &[LocationHistoryEntry],
    ) -> Result<(), LocationRepositoryError>;

    /// Get location history for a courier within a time range
    ///
    /// Returns entries ordered by timestamp ascending.
    async fn get_history(
        &self,
        courier_id: Uuid,
        time_range: TimeRange,
    ) -> Result<Vec<LocationHistoryEntry>, LocationRepositoryError>;

    /// Get location history for a courier with pagination
    ///
    /// Returns entries ordered by timestamp descending (most recent first).
    async fn get_history_paginated(
        &self,
        courier_id: Uuid,
        time_range: TimeRange,
        limit: u32,
        offset: u32,
    ) -> Result<Vec<LocationHistoryEntry>, LocationRepositoryError>;

    /// Get the last known location for a courier
    ///
    /// Returns the most recent location entry or None if no history exists.
    async fn get_last_location(
        &self,
        courier_id: Uuid,
    ) -> Result<Option<LocationHistoryEntry>, LocationRepositoryError>;

    /// Get last locations for multiple couriers
    ///
    /// Returns a map of courier_id to their last known location.
    async fn get_last_locations(
        &self,
        courier_ids: &[Uuid],
    ) -> Result<Vec<LocationHistoryEntry>, LocationRepositoryError>;

    /// Count location entries for a courier in a time range
    async fn count_history(
        &self,
        courier_id: Uuid,
        time_range: TimeRange,
    ) -> Result<u64, LocationRepositoryError>;

    /// Delete old location history
    ///
    /// Removes entries older than the specified number of days.
    /// Used for data retention policy.
    async fn delete_old_history(&self, older_than_days: u32) -> Result<u64, LocationRepositoryError>;
}
