//! Courier Repository Port
//!
//! Defines the interface for persisting and retrieving Courier aggregates.
//! This port is implemented by infrastructure adapters (e.g., PostgreSQL).

use async_trait::async_trait;
#[cfg(test)]
use mockall::automock;
use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::courier::{Courier, CourierStatus};

/// Filter criteria for querying couriers
#[derive(Debug, Clone, Default)]
pub struct CourierFilter {
    /// Filter by work zone
    pub work_zone: Option<String>,
    /// Filter by status (Free, Busy, Unavailable, Archived)
    pub status: Option<CourierStatus>,
    /// Filter by archived flag (true = only archived, false = only non-archived)
    pub archived: Option<bool>,
}

impl CourierFilter {
    /// Create a filter for a specific work zone
    pub fn by_work_zone(zone: &str) -> Self {
        Self {
            work_zone: Some(zone.to_string()),
            ..Default::default()
        }
    }

    /// Create a filter for a specific status
    pub fn with_status(mut self, status: CourierStatus) -> Self {
        self.status = Some(status);
        self
    }

    /// Filter for available couriers (Free) in a zone
    pub fn available_in_zone(zone: &str) -> Self {
        Self {
            work_zone: Some(zone.to_string()),
            status: Some(CourierStatus::Free),
            ..Default::default()
        }
    }

    /// Exclude archived couriers
    pub fn not_archived(mut self) -> Self {
        self.archived = Some(false);
        self
    }
}

/// Errors that can occur during repository operations
#[derive(Debug, Error)]
pub enum RepositoryError {
    /// Entity not found
    #[error("Courier not found: {0}")]
    NotFound(Uuid),

    /// Duplicate entry (email or phone already exists)
    #[error("Duplicate entry: {0}")]
    DuplicateEntry(String),

    /// Version conflict (optimistic locking failure)
    #[error("Version conflict: expected {expected}, found {actual}")]
    VersionConflict { expected: u32, actual: u32 },

    /// Database connection error
    #[error("Connection error: {0}")]
    ConnectionError(String),

    /// Query execution error
    #[error("Query error: {0}")]
    QueryError(String),

    /// Serialization/deserialization error
    #[error("Serialization error: {0}")]
    SerializationError(String),
}

/// Courier Repository Port
///
/// Defines the contract for storing and retrieving Courier aggregates.
/// Implementations handle the actual persistence mechanism (PostgreSQL, etc.).
#[cfg_attr(test, automock)]
#[async_trait]
pub trait CourierRepository: Send + Sync {
    /// Save a courier (insert or update)
    ///
    /// If the courier doesn't exist, it will be inserted.
    /// If it exists, it will be updated with optimistic locking.
    async fn save(&self, courier: &Courier) -> Result<(), RepositoryError>;

    /// Find a courier by ID
    async fn find_by_id(&self, id: Uuid) -> Result<Option<Courier>, RepositoryError>;

    /// Find a courier by phone number
    async fn find_by_phone(&self, phone: &str) -> Result<Option<Courier>, RepositoryError>;

    /// Find a courier by email
    async fn find_by_email(&self, email: &str) -> Result<Option<Courier>, RepositoryError>;

    /// Find couriers by work zone
    async fn find_by_work_zone(&self, zone: &str) -> Result<Vec<Courier>, RepositoryError>;

    /// Check if email is already registered
    async fn email_exists(&self, email: &str) -> Result<bool, RepositoryError>;

    /// Check if phone is already registered
    async fn phone_exists(&self, phone: &str) -> Result<bool, RepositoryError>;

    /// Delete a courier by ID
    async fn delete(&self, id: Uuid) -> Result<(), RepositoryError>;

    /// Archive a courier (soft delete)
    ///
    /// Sets the courier as archived in the database.
    /// The courier data is retained but marked as inactive.
    async fn archive(&self, id: Uuid) -> Result<(), RepositoryError>;

    /// List all couriers with pagination
    ///
    /// Returns couriers ordered by created_at descending.
    /// Use limit and offset for pagination.
    async fn list(&self, limit: u64, offset: u64) -> Result<Vec<Courier>, RepositoryError>;

    /// Find couriers matching the filter with pagination
    ///
    /// Returns couriers ordered by created_at descending.
    async fn find_by_filter(
        &self,
        filter: CourierFilter,
        limit: u64,
        offset: u64,
    ) -> Result<Vec<Courier>, RepositoryError>;

    /// Count couriers matching the filter
    async fn count_by_filter(&self, filter: CourierFilter) -> Result<u64, RepositoryError>;
}
