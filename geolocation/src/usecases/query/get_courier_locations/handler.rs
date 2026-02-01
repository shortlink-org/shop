//! Get Courier Locations Handler
//!
//! Fetches current locations for multiple couriers, using cache with DB fallback.

use std::collections::HashMap;
use std::sync::Arc;

use async_trait::async_trait;
use chrono::{DateTime, Utc};
use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::CourierLocation;
use crate::domain::ports::{
    CacheError, LocationCache, LocationRepository, QueryHandler, RepositoryError,
};

use super::query::MAX_COURIER_IDS;
use super::Query;

/// Errors that can occur during location fetch
#[derive(Debug, Error)]
pub enum GetCourierLocationsError {
    #[error("Invalid request: {0}")]
    InvalidRequest(String),

    #[error("Too many courier IDs: {0} (max: {MAX_COURIER_IDS})")]
    TooManyRequests(usize),

    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    #[error("Cache error: {0}")]
    CacheError(#[from] CacheError),
}

/// Location info for a single courier
#[derive(Debug, Clone)]
pub struct LocationInfo {
    /// The location data (if found)
    pub location: Option<CourierLocation>,
    /// Last update time (if found)
    pub last_updated: Option<DateTime<Utc>>,
    /// Whether the location was found
    pub found: bool,
}

impl LocationInfo {
    fn found(location: CourierLocation) -> Self {
        let last_updated = location.updated_at();
        Self {
            location: Some(location),
            last_updated: Some(last_updated),
            found: true,
        }
    }

    fn not_found() -> Self {
        Self {
            location: None,
            last_updated: None,
            found: false,
        }
    }
}

/// Response containing locations for all requested couriers
#[derive(Debug, Clone)]
pub struct Response {
    /// Map of courier_id -> location info
    pub locations: HashMap<Uuid, LocationInfo>,
}

/// Get Courier Locations Handler
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
impl<R, C> QueryHandler<Query, Response> for Handler<R, C>
where
    R: LocationRepository + Send + Sync,
    C: LocationCache + Send + Sync,
{
    type Error = GetCourierLocationsError;

    async fn handle(&self, query: Query) -> Result<Response, Self::Error> {
        // 1. Validate query
        query.validate().map_err(GetCourierLocationsError::InvalidRequest)?;

        if query.courier_ids.len() > MAX_COURIER_IDS {
            return Err(GetCourierLocationsError::TooManyRequests(query.courier_ids.len()));
        }

        let mut result: HashMap<Uuid, LocationInfo> = HashMap::new();
        let mut uncached_ids: Vec<Uuid> = Vec::new();

        // 2. Try to get from cache first
        let cached = self.location_cache.get_locations(&query.courier_ids).await?;

        for (courier_id, maybe_location) in cached {
            match maybe_location {
                Some(location) => {
                    result.insert(courier_id, LocationInfo::found(location));
                }
                None => {
                    uncached_ids.push(courier_id);
                }
            }
        }

        // 3. Fetch uncached from database
        if !uncached_ids.is_empty() {
            let db_locations = self.location_repo.get_current_locations(&uncached_ids).await?;

            // Create a map for quick lookup
            let db_map: HashMap<Uuid, CourierLocation> = db_locations
                .into_iter()
                .map(|loc| (loc.courier_id(), loc))
                .collect();

            // Update results and cache
            for courier_id in uncached_ids {
                if let Some(location) = db_map.get(&courier_id) {
                    // Update cache with DB result
                    let _ = self.location_cache.set_location(location).await;
                    result.insert(courier_id, LocationInfo::found(location.clone()));
                } else {
                    result.insert(courier_id, LocationInfo::not_found());
                }
            }
        }

        Ok(Response { locations: result })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::{Location, LocationHistoryEntry, TimeRange};
    use async_trait::async_trait;
    use chrono::Utc;
    use std::sync::Mutex;

    // === Mock Repository ===

    struct MockLocationRepository {
        locations: Mutex<HashMap<Uuid, CourierLocation>>,
    }

    impl MockLocationRepository {
        fn new() -> Self {
            Self {
                locations: Mutex::new(HashMap::new()),
            }
        }

        fn add_location(&self, location: CourierLocation) {
            let mut locations = self.locations.lock().unwrap();
            locations.insert(location.courier_id(), location);
        }
    }

    #[async_trait]
    impl LocationRepository for MockLocationRepository {
        async fn save_current_location(&self, location: &CourierLocation) -> Result<(), RepositoryError> {
            let mut locations = self.locations.lock().unwrap();
            locations.insert(location.courier_id(), location.clone());
            Ok(())
        }

        async fn get_current_location(&self, courier_id: Uuid) -> Result<Option<CourierLocation>, RepositoryError> {
            let locations = self.locations.lock().unwrap();
            Ok(locations.get(&courier_id).cloned())
        }

        async fn get_current_locations(&self, courier_ids: &[Uuid]) -> Result<Vec<CourierLocation>, RepositoryError> {
            let locations = self.locations.lock().unwrap();
            Ok(courier_ids
                .iter()
                .filter_map(|id| locations.get(id).cloned())
                .collect())
        }

        async fn save_history_entry(&self, _entry: &LocationHistoryEntry) -> Result<(), RepositoryError> {
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

    // === Test Helpers ===

    fn create_test_courier_location(courier_id: Uuid) -> CourierLocation {
        let location = Location::from_stored(52.52, 13.405, 15.0, Utc::now(), Some(35.0), Some(180.0)).unwrap();
        CourierLocation::new(courier_id, location)
    }

    // === Tests ===

    #[tokio::test]
    async fn test_get_single_courier_location() {
        let repo = Arc::new(MockLocationRepository::new());
        let cache = Arc::new(MockLocationCache::new());

        let courier_id = Uuid::new_v4();
        let location = create_test_courier_location(courier_id);
        repo.add_location(location);

        let handler = Handler::new(repo, cache);
        let query = Query::single(courier_id);

        let result = handler.handle(query).await;
        assert!(result.is_ok());

        let response = result.unwrap();
        assert!(response.locations.contains_key(&courier_id));
        assert!(response.locations[&courier_id].found);
    }

    #[tokio::test]
    async fn test_get_multiple_courier_locations() {
        let repo = Arc::new(MockLocationRepository::new());
        let cache = Arc::new(MockLocationCache::new());

        let courier_id1 = Uuid::new_v4();
        let courier_id2 = Uuid::new_v4();
        let courier_id3 = Uuid::new_v4(); // This one won't have a location

        repo.add_location(create_test_courier_location(courier_id1));
        repo.add_location(create_test_courier_location(courier_id2));

        let handler = Handler::new(repo, cache);
        let query = Query::new(vec![courier_id1, courier_id2, courier_id3]);

        let result = handler.handle(query).await;
        assert!(result.is_ok());

        let response = result.unwrap();
        assert!(response.locations[&courier_id1].found);
        assert!(response.locations[&courier_id2].found);
        assert!(!response.locations[&courier_id3].found);
    }

    #[tokio::test]
    async fn test_cache_hit() {
        let repo = Arc::new(MockLocationRepository::new());
        let cache = Arc::new(MockLocationCache::new());

        let courier_id = Uuid::new_v4();
        let location = create_test_courier_location(courier_id);

        // Put in cache only
        cache.set_location(&location).await.unwrap();

        let handler = Handler::new(repo, cache);
        let query = Query::single(courier_id);

        let result = handler.handle(query).await;
        assert!(result.is_ok());
        assert!(result.unwrap().locations[&courier_id].found);
    }

    #[tokio::test]
    async fn test_empty_query() {
        let repo = Arc::new(MockLocationRepository::new());
        let cache = Arc::new(MockLocationCache::new());
        let handler = Handler::new(repo, cache);

        let query = Query::new(vec![]);
        let result = handler.handle(query).await;
        assert!(matches!(result, Err(GetCourierLocationsError::InvalidRequest(_))));
    }
}
