//! Geolocation Service Port
//!
//! Defines the interface for recording and reading courier locations.
//! This port provides a single integration point for "where is the courier now"
//! used by dispatch, deliver_order, pick_up_order, and get_courier/get_pool.

use async_trait::async_trait;
use chrono::{DateTime, Utc};
use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::{CourierLocation, CourierLocationError};
use crate::domain::model::vo::location::Location;

/// Errors for geolocation operations
#[derive(Debug, Error)]
pub enum GeolocationServiceError {
    #[error("Invalid location: {0}")]
    InvalidLocation(#[from] CourierLocationError),

    #[error("Cache error: {0}")]
    CacheError(String),

    #[error("Repository error: {0}")]
    RepositoryError(String),
}

/// Geolocation Service Port
///
/// Single port for updating and reading courier current location.
/// Implementations delegate to location cache and/or location repository.
#[async_trait]
pub trait GeolocationService: Send + Sync {
    /// Record courier location (e.g. after pickup, after delivery, or from UpdateLocation use case).
    async fn update_location(
        &self,
        courier_id: Uuid,
        location: Location,
        timestamp: DateTime<Utc>,
    ) -> Result<(), GeolocationServiceError>;

    /// Get current location for a courier, if available.
    async fn get_location(
        &self,
        courier_id: Uuid,
    ) -> Result<Option<CourierLocation>, GeolocationServiceError>;
}
