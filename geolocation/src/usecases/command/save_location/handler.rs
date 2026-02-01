//! Save Location Handler
//!
//! Handles saving a courier's location to database and cache.

use std::sync::Arc;

use chrono::{DateTime, Utc};
use thiserror::Error;
use uuid::Uuid;

use async_trait::async_trait;

use crate::domain::model::{CourierLocation, Location, LocationError, LocationHistoryEntry};
use crate::domain::ports::{
    CacheError, CommandHandlerWithResult, LocationCache, LocationRepository, RepositoryError,
};

use super::Command;

/// Errors that can occur during location save
#[derive(Debug, Error)]
pub enum SaveLocationError {
    #[error("Invalid location: {0}")]
    InvalidLocation(#[from] LocationError),

    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    #[error("Cache error: {0}")]
    CacheError(#[from] CacheError),
}

/// Response from saving a location
#[derive(Debug, Clone)]
pub struct Response {
    /// Courier ID
    pub courier_id: Uuid,
    /// When the location was updated
    pub updated_at: DateTime<Utc>,
    /// History entry ID
    pub location_id: Uuid,
    /// Success flag
    pub success: bool,
}

/// Save Location Handler
pub struct Handler<R, C>
where
    R: LocationRepository,
    C: LocationCache,
{
    location_repo: Arc<R>,
    location_cache: Arc<C>,
}

impl<R, C> Handler<R, C>
where
    R: LocationRepository,
    C: LocationCache,
{
    /// Create a new handler
    pub fn new(location_repo: Arc<R>, location_cache: Arc<C>) -> Self {
        Self {
            location_repo,
            location_cache,
        }
    }
}

#[async_trait]
impl<R, C> CommandHandlerWithResult<Command, Response> for Handler<R, C>
where
    R: LocationRepository + Send + Sync,
    C: LocationCache + Send + Sync,
{
    type Error = SaveLocationError;

    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        // 1. Create and validate Location value object
        let location = Location::new(
            cmd.latitude,
            cmd.longitude,
            cmd.accuracy,
            cmd.timestamp,
            cmd.speed,
            cmd.heading,
        )?;

        // 2. Create CourierLocation entity
        let courier_location = CourierLocation::new(cmd.courier_id, location.clone());

        // 3. Create history entry
        let history_entry = LocationHistoryEntry::new(cmd.courier_id, location);
        let location_id = history_entry.id();

        // 4. Save current location to repository
        self.location_repo.save_current_location(&courier_location).await?;

        // 5. Save history entry
        self.location_repo.save_history_entry(&history_entry).await?;

        // 6. Update cache
        self.location_cache.set_location(&courier_location).await?;

        let updated_at = courier_location.updated_at();

        Ok(Response {
            courier_id: cmd.courier_id,
            updated_at,
            location_id,
            success: true,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::TimeRange;
    use async_trait::async_trait;
    use chrono::Utc;
    use std::collections::HashMap;
    use std::sync::Mutex;

    // === Mock Repository ===

    struct MockLocationRepository {
        current_locations: Mutex<HashMap<Uuid, CourierLocation>>,
        history: Mutex<Vec<LocationHistoryEntry>>,
    }

    impl MockLocationRepository {
        fn new() -> Self {
            Self {
                current_locations: Mutex::new(HashMap::new()),
                history: Mutex::new(Vec::new()),
            }
        }
    }

    #[async_trait]
    impl LocationRepository for MockLocationRepository {
        async fn save_current_location(&self, location: &CourierLocation) -> Result<(), RepositoryError> {
            let mut locations = self.current_locations.lock().unwrap();
            locations.insert(location.courier_id(), location.clone());
            Ok(())
        }

        async fn get_current_location(&self, courier_id: Uuid) -> Result<Option<CourierLocation>, RepositoryError> {
            let locations = self.current_locations.lock().unwrap();
            Ok(locations.get(&courier_id).cloned())
        }

        async fn get_current_locations(&self, courier_ids: &[Uuid]) -> Result<Vec<CourierLocation>, RepositoryError> {
            let locations = self.current_locations.lock().unwrap();
            Ok(courier_ids
                .iter()
                .filter_map(|id| locations.get(id).cloned())
                .collect())
        }

        async fn save_history_entry(&self, entry: &LocationHistoryEntry) -> Result<(), RepositoryError> {
            let mut history = self.history.lock().unwrap();
            history.push(entry.clone());
            Ok(())
        }

        async fn get_location_history(
            &self,
            _courier_id: Uuid,
            _time_range: &TimeRange,
            _limit: Option<u32>,
        ) -> Result<Vec<LocationHistoryEntry>, RepositoryError> {
            Ok(vec![])
        }

        async fn delete_history_before(&self, _before: chrono::DateTime<Utc>) -> Result<u64, RepositoryError> {
            Ok(0)
        }
    }

    // === Mock Cache ===

    struct MockLocationCache {
        cache: Mutex<HashMap<Uuid, CourierLocation>>,
    }

    impl MockLocationCache {
        fn new() -> Self {
            Self {
                cache: Mutex::new(HashMap::new()),
            }
        }
    }

    #[async_trait]
    impl LocationCache for MockLocationCache {
        async fn set_location(&self, location: &CourierLocation) -> Result<(), CacheError> {
            let mut cache = self.cache.lock().unwrap();
            cache.insert(location.courier_id(), location.clone());
            Ok(())
        }

        async fn get_location(&self, courier_id: Uuid) -> Result<Option<CourierLocation>, CacheError> {
            let cache = self.cache.lock().unwrap();
            Ok(cache.get(&courier_id).cloned())
        }

        async fn get_locations(&self, courier_ids: &[Uuid]) -> Result<Vec<(Uuid, Option<CourierLocation>)>, CacheError> {
            let cache = self.cache.lock().unwrap();
            Ok(courier_ids
                .iter()
                .map(|id| (*id, cache.get(id).cloned()))
                .collect())
        }

        async fn remove_location(&self, courier_id: Uuid) -> Result<(), CacheError> {
            let mut cache = self.cache.lock().unwrap();
            cache.remove(&courier_id);
            Ok(())
        }

        async fn exists(&self, courier_id: Uuid) -> Result<bool, CacheError> {
            let cache = self.cache.lock().unwrap();
            Ok(cache.contains_key(&courier_id))
        }
    }

    // === Tests ===

    #[tokio::test]
    async fn test_save_location_success() {
        let repo = Arc::new(MockLocationRepository::new());
        let cache = Arc::new(MockLocationCache::new());
        let handler = Handler::new(repo.clone(), cache.clone());

        let courier_id = Uuid::new_v4();
        let cmd = Command::new(
            courier_id,
            52.52,
            13.405,
            15.0,
            Utc::now(),
            Some(35.0),
            Some(180.0),
        );

        let result = handler.handle(cmd).await;
        assert!(result.is_ok(), "Expected Ok, got: {:?}", result.err());

        let response = result.unwrap();
        assert_eq!(response.courier_id, courier_id);
        assert!(response.success);

        // Verify saved to repository
        let saved = repo.get_current_location(courier_id).await.unwrap();
        assert!(saved.is_some());
        assert_eq!(saved.unwrap().latitude(), 52.52);

        // Verify saved to cache
        let cached = cache.get_location(courier_id).await.unwrap();
        assert!(cached.is_some());
    }

    #[tokio::test]
    async fn test_save_location_invalid_coordinates() {
        let repo = Arc::new(MockLocationRepository::new());
        let cache = Arc::new(MockLocationCache::new());
        let handler = Handler::new(repo, cache);

        let cmd = Command::new(
            Uuid::new_v4(),
            91.0, // Invalid latitude
            13.405,
            15.0,
            Utc::now(),
            None,
            None,
        );

        let result = handler.handle(cmd).await;
        assert!(matches!(result, Err(SaveLocationError::InvalidLocation(_))));
    }

    #[tokio::test]
    async fn test_save_location_invalid_accuracy() {
        let repo = Arc::new(MockLocationRepository::new());
        let cache = Arc::new(MockLocationCache::new());
        let handler = Handler::new(repo, cache);

        let cmd = Command::new(
            Uuid::new_v4(),
            52.52,
            13.405,
            0.0, // Invalid accuracy
            Utc::now(),
            None,
            None,
        );

        let result = handler.handle(cmd).await;
        assert!(matches!(result, Err(SaveLocationError::InvalidLocation(_))));
    }
}
