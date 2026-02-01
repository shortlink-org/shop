//! Courier Repository Port
//!
//! Defines the interface for persisting and retrieving Courier aggregates.
//! This port is implemented by infrastructure adapters (e.g., PostgreSQL).

use async_trait::async_trait;
#[cfg(test)]
use mockall::automock;
use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::courier::Courier;

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
}
