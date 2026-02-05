//! Geolocation Service stub implementation.
//!
//! Delegates to LocationCache (hot path) and LocationRepository (history).
//! Provides a single port for recording and reading courier locations.

use std::sync::Arc;

use async_trait::async_trait;
use chrono::{DateTime, Utc};

use crate::domain::model::{CourierLocation, LocationHistoryEntry};
use crate::domain::model::vo::location::Location;
use crate::domain::ports::{
    GeolocationService, GeolocationServiceError, LocationCache, LocationRepository,
};

/// Stub implementation of GeolocationService using existing cache and repository.
pub struct StubGeolocationService<LC, LR>
where
    LC: LocationCache,
    LR: LocationRepository,
{
    location_cache: Arc<LC>,
    location_repository: Arc<LR>,
}

impl<LC, LR> StubGeolocationService<LC, LR>
where
    LC: LocationCache,
    LR: LocationRepository,
{
    pub fn new(location_cache: Arc<LC>, location_repository: Arc<LR>) -> Self {
        Self {
            location_cache,
            location_repository,
        }
    }
}

#[async_trait]
impl<LC, LR> GeolocationService for StubGeolocationService<LC, LR>
where
    LC: LocationCache + Send + Sync,
    LR: LocationRepository + Send + Sync,
{
    async fn update_location(
        &self,
        courier_id: uuid::Uuid,
        location: Location,
        timestamp: DateTime<Utc>,
    ) -> Result<(), GeolocationServiceError> {
        let courier_location = CourierLocation::new(courier_id, location, timestamp, None, None)?;
        self.location_cache
            .set_location(&courier_location)
            .await
            .map_err(|e| GeolocationServiceError::CacheError(e.to_string()))?;
        let entry = LocationHistoryEntry::new(courier_id, location, timestamp, None, None);
        self.location_repository
            .save(&entry)
            .await
            .map_err(|e| GeolocationServiceError::RepositoryError(e.to_string()))?;
        Ok(())
    }

    async fn get_location(
        &self,
        courier_id: uuid::Uuid,
    ) -> Result<Option<CourierLocation>, GeolocationServiceError> {
        if let Ok(Some(loc)) = self.location_cache.get_location(courier_id).await {
            return Ok(Some(loc));
        }
        if let Ok(Some(entry)) = self.location_repository.get_last_location(courier_id).await {
            let loc = CourierLocation::from_stored(
                entry.courier_id(),
                *entry.location(),
                entry.timestamp(),
                entry.speed(),
                entry.heading(),
            )
            .map_err(GeolocationServiceError::InvalidLocation)?;
            return Ok(Some(loc));
        }
        Ok(None)
    }
}
