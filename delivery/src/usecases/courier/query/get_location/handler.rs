//! Get Courier Location Handler
//!
//! Handles retrieving a courier's current location.
//!
//! ## Flow
//! 1. Check cache for current location (hot data)
//! 2. If not in cache, check repository for last known location
//! 3. Return location or None

use std::sync::Arc;

use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::CourierLocation;
use crate::domain::ports::{
    LocationCache, LocationCacheError, LocationRepository, LocationRepositoryError, QueryHandler,
};

use super::{BatchQuery, Query};

/// Errors that can occur during location retrieval
#[derive(Debug, Error)]
pub enum GetLocationError {
    /// Cache error
    #[error("Cache error: {0}")]
    CacheError(#[from] LocationCacheError),

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] LocationRepositoryError),
}

/// Response from get location query
#[derive(Debug, Clone)]
pub struct Response {
    /// Courier ID
    pub courier_id: Uuid,
    /// Current location (None if no location data available)
    pub location: Option<CourierLocation>,
    /// Whether the location came from cache (true) or repository (false)
    pub from_cache: bool,
}

/// Response from batch get locations query
#[derive(Debug, Clone)]
pub struct BatchResponse {
    /// Locations for each courier
    pub locations: Vec<Response>,
}

/// Get Courier Location Handler
pub struct Handler<LC, LR>
where
    LC: LocationCache,
    LR: LocationRepository,
{
    location_cache: Arc<LC>,
    location_repository: Arc<LR>,
}

impl<LC, LR> Handler<LC, LR>
where
    LC: LocationCache,
    LR: LocationRepository,
{
    /// Create a new handler instance
    pub fn new(location_cache: Arc<LC>, location_repository: Arc<LR>) -> Self {
        Self {
            location_cache,
            location_repository,
        }
    }

    /// Get location from cache or repository
    async fn get_location_for_courier(&self, courier_id: Uuid) -> Result<Response, GetLocationError> {
        // 1. Try cache first (hot data)
        if let Some(location) = self.location_cache.get_location(courier_id).await? {
            return Ok(Response {
                courier_id,
                location: Some(location),
                from_cache: true,
            });
        }

        // 2. Fall back to repository (last known location)
        if let Some(history_entry) = self.location_repository.get_last_location(courier_id).await? {
            // Convert history entry to CourierLocation
            let location = CourierLocation::from_stored(
                history_entry.courier_id(),
                *history_entry.location(),
                history_entry.timestamp(),
                history_entry.speed(),
                history_entry.heading(),
            )
            .ok();

            return Ok(Response {
                courier_id,
                location,
                from_cache: false,
            });
        }

        // 3. No location data available
        Ok(Response {
            courier_id,
            location: None,
            from_cache: false,
        })
    }
}

impl<LC, LR> QueryHandler<Query, Response> for Handler<LC, LR>
where
    LC: LocationCache + Send + Sync,
    LR: LocationRepository + Send + Sync,
{
    type Error = GetLocationError;

    /// Handle the GetCourierLocation query
    async fn handle(&self, query: Query) -> Result<Response, Self::Error> {
        self.get_location_for_courier(query.courier_id).await
    }
}

impl<LC, LR> QueryHandler<BatchQuery, BatchResponse> for Handler<LC, LR>
where
    LC: LocationCache + Send + Sync,
    LR: LocationRepository + Send + Sync,
{
    type Error = GetLocationError;

    /// Handle the batch GetCourierLocations query
    async fn handle(&self, query: BatchQuery) -> Result<BatchResponse, Self::Error> {
        let mut locations = Vec::with_capacity(query.courier_ids.len());

        for courier_id in query.courier_ids {
            let response = self.get_location_for_courier(courier_id).await?;
            locations.push(response);
        }

        Ok(BatchResponse { locations })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::courier_location::{LocationHistoryEntry, TimeRange};
    use crate::domain::model::vo::location::Location;
    use async_trait::async_trait;
    use chrono::Utc;
    use std::collections::HashMap;
    use std::sync::Mutex;

    // Mock LocationCache
    struct MockLocationCache {
        locations: Mutex<HashMap<Uuid, CourierLocation>>,
    }

    impl MockLocationCache {
        fn new() -> Self {
            Self {
                locations: Mutex::new(HashMap::new()),
            }
        }

        fn set(&self, location: CourierLocation) {
            self.locations
                .lock()
                .unwrap()
                .insert(location.courier_id(), location);
        }
    }

    #[async_trait]
    impl LocationCache for MockLocationCache {
        async fn set_location(&self, location: &CourierLocation) -> Result<(), LocationCacheError> {
            self.locations
                .lock()
                .unwrap()
                .insert(location.courier_id(), location.clone());
            Ok(())
        }

        async fn get_location(&self, courier_id: Uuid) -> Result<Option<CourierLocation>, LocationCacheError> {
            Ok(self.locations.lock().unwrap().get(&courier_id).cloned())
        }

        async fn get_locations(&self, courier_ids: &[Uuid]) -> Result<Vec<CourierLocation>, LocationCacheError> {
            let locations = self.locations.lock().unwrap();
            Ok(courier_ids
                .iter()
                .filter_map(|id| locations.get(id).cloned())
                .collect())
        }

        async fn delete_location(&self, courier_id: Uuid) -> Result<(), LocationCacheError> {
            self.locations.lock().unwrap().remove(&courier_id);
            Ok(())
        }

        async fn has_location(&self, courier_id: Uuid) -> Result<bool, LocationCacheError> {
            Ok(self.locations.lock().unwrap().contains_key(&courier_id))
        }

        async fn get_all_locations(&self) -> Result<Vec<CourierLocation>, LocationCacheError> {
            Ok(self.locations.lock().unwrap().values().cloned().collect())
        }

        async fn get_active_courier_ids(&self) -> Result<Vec<Uuid>, LocationCacheError> {
            Ok(self.locations.lock().unwrap().keys().cloned().collect())
        }
    }

    // Mock LocationRepository
    struct MockLocationRepository {
        entries: Mutex<Vec<LocationHistoryEntry>>,
    }

    impl MockLocationRepository {
        fn new() -> Self {
            Self {
                entries: Mutex::new(Vec::new()),
            }
        }

        fn add(&self, entry: LocationHistoryEntry) {
            self.entries.lock().unwrap().push(entry);
        }
    }

    #[async_trait]
    impl LocationRepository for MockLocationRepository {
        async fn save(&self, entry: &LocationHistoryEntry) -> Result<(), LocationRepositoryError> {
            self.entries.lock().unwrap().push(entry.clone());
            Ok(())
        }

        async fn save_batch(&self, entries: &[LocationHistoryEntry]) -> Result<(), LocationRepositoryError> {
            self.entries.lock().unwrap().extend(entries.iter().cloned());
            Ok(())
        }

        async fn get_history(
            &self,
            courier_id: Uuid,
            time_range: TimeRange,
        ) -> Result<Vec<LocationHistoryEntry>, LocationRepositoryError> {
            Ok(self
                .entries
                .lock()
                .unwrap()
                .iter()
                .filter(|e| e.courier_id() == courier_id && time_range.contains(e.timestamp()))
                .cloned()
                .collect())
        }

        async fn get_history_paginated(
            &self,
            courier_id: Uuid,
            time_range: TimeRange,
            limit: u32,
            offset: u32,
        ) -> Result<Vec<LocationHistoryEntry>, LocationRepositoryError> {
            let all = self.get_history(courier_id, time_range).await?;
            Ok(all
                .into_iter()
                .skip(offset as usize)
                .take(limit as usize)
                .collect())
        }

        async fn get_last_location(
            &self,
            courier_id: Uuid,
        ) -> Result<Option<LocationHistoryEntry>, LocationRepositoryError> {
            Ok(self
                .entries
                .lock()
                .unwrap()
                .iter()
                .filter(|e| e.courier_id() == courier_id)
                .max_by_key(|e| e.timestamp())
                .cloned())
        }

        async fn get_last_locations(
            &self,
            courier_ids: &[Uuid],
        ) -> Result<Vec<LocationHistoryEntry>, LocationRepositoryError> {
            let mut results = Vec::new();
            for id in courier_ids {
                if let Some(entry) = self.get_last_location(*id).await? {
                    results.push(entry);
                }
            }
            Ok(results)
        }

        async fn count_history(
            &self,
            courier_id: Uuid,
            time_range: TimeRange,
        ) -> Result<u64, LocationRepositoryError> {
            Ok(self.get_history(courier_id, time_range).await?.len() as u64)
        }

        async fn delete_old_history(&self, _older_than_days: u32) -> Result<u64, LocationRepositoryError> {
            Ok(0)
        }
    }

    #[tokio::test]
    async fn test_get_location_from_cache() {
        let courier_id = Uuid::new_v4();
        let cache = Arc::new(MockLocationCache::new());
        let repo = Arc::new(MockLocationRepository::new());

        // Add location to cache
        let location = CourierLocation::new(
            courier_id,
            Location::new(52.52, 13.405, 10.0).unwrap(),
            Utc::now(),
            Some(35.0),
            Some(180.0),
        )
        .unwrap();
        cache.set(location);

        let handler = Handler::new(cache, repo);
        let query = Query::new(courier_id);
        let result = handler.handle(query).await;

        assert!(result.is_ok());
        let response = result.unwrap();
        assert!(response.location.is_some());
        assert!(response.from_cache);
        assert_eq!(response.location.unwrap().latitude(), 52.52);
    }

    #[tokio::test]
    async fn test_get_location_from_repository() {
        let courier_id = Uuid::new_v4();
        let cache = Arc::new(MockLocationCache::new());
        let repo = Arc::new(MockLocationRepository::new());

        // Add location to repository (not cache)
        let entry = LocationHistoryEntry::new(
            courier_id,
            Location::new(48.1351, 11.582, 10.0).unwrap(),
            Utc::now() - chrono::Duration::minutes(1),
            Some(50.0),
            Some(90.0),
        );
        repo.add(entry);

        let handler = Handler::new(cache, repo);
        let query = Query::new(courier_id);
        let result = handler.handle(query).await;

        assert!(result.is_ok());
        let response = result.unwrap();
        assert!(response.location.is_some());
        assert!(!response.from_cache);
    }

    #[tokio::test]
    async fn test_get_location_not_found() {
        let cache = Arc::new(MockLocationCache::new());
        let repo = Arc::new(MockLocationRepository::new());

        let handler = Handler::new(cache, repo);
        let query = Query::new(Uuid::new_v4());
        let result = handler.handle(query).await;

        assert!(result.is_ok());
        let response = result.unwrap();
        assert!(response.location.is_none());
    }
}
