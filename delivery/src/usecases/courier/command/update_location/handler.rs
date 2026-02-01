//! Update Courier Location Handler
//!
//! Handles updating a courier's GPS location in real-time.
//!
//! ## Flow
//! 1. Validate courier exists
//! 2. Create CourierLocation entity with validation
//! 3. Save to location cache (Redis - hot data)
//! 4. Save to location history (PostgreSQL - cold data)

use std::sync::Arc;

use thiserror::Error;

use crate::domain::model::{CourierLocation, CourierLocationError, LocationHistoryEntry};
use crate::domain::ports::{
    CommandHandler, CourierRepository, LocationCache, LocationCacheError, LocationRepository,
    LocationRepositoryError, RepositoryError,
};

use super::Command;

/// Errors that can occur during courier location update
#[derive(Debug, Error)]
pub enum UpdateCourierLocationError {
    /// Courier not found
    #[error("Courier not found: {0}")]
    CourierNotFound(String),

    /// Invalid location data
    #[error("Invalid location data: {0}")]
    InvalidLocation(#[from] CourierLocationError),

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    /// Location repository error
    #[error("Location repository error: {0}")]
    LocationRepositoryError(#[from] LocationRepositoryError),

    /// Cache error
    #[error("Cache error: {0}")]
    CacheError(#[from] LocationCacheError),
}

/// Update Courier Location Handler
pub struct Handler<R, LC, LR>
where
    R: CourierRepository,
    LC: LocationCache,
    LR: LocationRepository,
{
    courier_repository: Arc<R>,
    location_cache: Arc<LC>,
    location_repository: Arc<LR>,
}

impl<R, LC, LR> Handler<R, LC, LR>
where
    R: CourierRepository,
    LC: LocationCache,
    LR: LocationRepository,
{
    /// Create a new handler instance
    pub fn new(
        courier_repository: Arc<R>,
        location_cache: Arc<LC>,
        location_repository: Arc<LR>,
    ) -> Self {
        Self {
            courier_repository,
            location_cache,
            location_repository,
        }
    }
}

impl<R, LC, LR> CommandHandler<Command> for Handler<R, LC, LR>
where
    R: CourierRepository + Send + Sync,
    LC: LocationCache + Send + Sync,
    LR: LocationRepository + Send + Sync,
{
    type Error = UpdateCourierLocationError;

    /// Handle the UpdateCourierLocation command
    async fn handle(&self, cmd: Command) -> Result<(), Self::Error> {
        // 1. Validate courier exists
        let _courier = self
            .courier_repository
            .find_by_id(cmd.courier_id)
            .await?
            .ok_or_else(|| {
                UpdateCourierLocationError::CourierNotFound(cmd.courier_id.to_string())
            })?;

        // 2. Create and validate CourierLocation entity
        let courier_location = CourierLocation::new(
            cmd.courier_id,
            cmd.location,
            cmd.timestamp,
            cmd.speed,
            cmd.heading,
        )?;

        // 3. Save to location cache (Redis - for real-time queries)
        self.location_cache.set_location(&courier_location).await?;

        // 4. Save to location history (PostgreSQL - for analytics)
        let history_entry = LocationHistoryEntry::new(
            cmd.courier_id,
            cmd.location,
            cmd.timestamp,
            cmd.speed,
            cmd.heading,
        );
        self.location_repository.save(&history_entry).await?;

        // TODO: Publish CourierLocationUpdatedEvent to message broker
        // event_publisher.publish(CourierLocationUpdatedEvent {
        //     courier_id: cmd.courier_id,
        //     location: cmd.location,
        //     timestamp: cmd.timestamp,
        // }).await?;

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::vo::location::Location;
    use crate::domain::model::courier_location::TimeRange;
    use crate::domain::model::courier::{Courier, CourierStatus};
    use chrono::Utc;
    use std::collections::HashMap;
    use std::sync::Mutex;
    use uuid::Uuid;
    use async_trait::async_trait;

    // Mock CourierRepository
    struct MockCourierRepository {
        couriers: Mutex<HashMap<Uuid, bool>>,
    }

    impl MockCourierRepository {
        fn new() -> Self {
            Self {
                couriers: Mutex::new(HashMap::new()),
            }
        }

        fn add_courier(&self, id: Uuid) {
            self.couriers.lock().unwrap().insert(id, true);
        }
    }

    #[async_trait]
    impl crate::domain::ports::CourierRepository for MockCourierRepository {
        async fn save(&self, _courier: &Courier) -> Result<(), RepositoryError> {
            Ok(())
        }

        async fn find_by_id(&self, id: Uuid) -> Result<Option<Courier>, RepositoryError> {
            if self.couriers.lock().unwrap().contains_key(&id) {
                Ok(Some(create_mock_courier(id)))
            } else {
                Ok(None)
            }
        }

        async fn find_by_phone(&self, _phone: &str) -> Result<Option<Courier>, RepositoryError> {
            Ok(None)
        }

        async fn find_by_email(&self, _email: &str) -> Result<Option<Courier>, RepositoryError> {
            Ok(None)
        }

        async fn find_by_work_zone(&self, _zone: &str) -> Result<Vec<Courier>, RepositoryError> {
            Ok(vec![])
        }

        async fn email_exists(&self, _email: &str) -> Result<bool, RepositoryError> {
            Ok(false)
        }

        async fn phone_exists(&self, _phone: &str) -> Result<bool, RepositoryError> {
            Ok(false)
        }

        async fn delete(&self, _id: Uuid) -> Result<(), RepositoryError> {
            Ok(())
        }

        async fn archive(&self, _id: Uuid) -> Result<(), RepositoryError> {
            Ok(())
        }

        async fn list(&self, _limit: u64, _offset: u64) -> Result<Vec<Courier>, RepositoryError> {
            Ok(vec![])
        }
    }

    fn create_mock_courier(id: Uuid) -> Courier {
        use crate::domain::model::courier::WorkHours;
        use crate::domain::model::vo::TransportType;

        let work_hours = WorkHours::new(
            chrono::NaiveTime::from_hms_opt(0, 0, 0).unwrap(),
            chrono::NaiveTime::from_hms_opt(23, 59, 0).unwrap(),
            vec![1, 2, 3, 4, 5, 6, 7],
        )
        .unwrap();

        Courier::reconstitute(
            crate::domain::model::courier::CourierId::from_uuid(id),
            "Test Courier".to_string(),
            "+491234567890".to_string(),
            "test@example.com".to_string(),
            TransportType::Car,
            100.0,
            "*".to_string(),
            work_hours,
            None,
            CourierStatus::Free,
            crate::domain::model::courier::CourierCapacity::new(5),
            5.0,
            100,
            0,
            Utc::now(),
            Utc::now(),
            1,
        )
    }

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

        fn get(&self, id: Uuid) -> Option<CourierLocation> {
            self.locations.lock().unwrap().get(&id).cloned()
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

        fn count(&self) -> usize {
            self.entries.lock().unwrap().len()
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
                .filter(|e| {
                    e.courier_id() == courier_id
                        && time_range.contains(e.timestamp())
                })
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
    async fn test_update_location_success() {
        let courier_id = Uuid::new_v4();
        let courier_repo = Arc::new(MockCourierRepository::new());
        courier_repo.add_courier(courier_id);

        let location_cache = Arc::new(MockLocationCache::new());
        let location_repo = Arc::new(MockLocationRepository::new());

        let handler = Handler::new(courier_repo, location_cache.clone(), location_repo.clone());

        let location = Location::new(52.52, 13.405, 10.0).unwrap();
        let cmd = Command::new(courier_id, location, Utc::now(), Some(35.0), Some(180.0));

        let result = handler.handle(cmd).await;
        assert!(result.is_ok());

        // Verify location was cached
        let cached = location_cache.get(courier_id);
        assert!(cached.is_some());
        assert_eq!(cached.unwrap().latitude(), 52.52);

        // Verify location was saved to history
        assert_eq!(location_repo.count(), 1);
    }

    #[tokio::test]
    async fn test_update_location_courier_not_found() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let location_cache = Arc::new(MockLocationCache::new());
        let location_repo = Arc::new(MockLocationRepository::new());

        let handler = Handler::new(courier_repo, location_cache, location_repo);

        let location = Location::new(52.52, 13.405, 10.0).unwrap();
        let cmd = Command::new(Uuid::new_v4(), location, Utc::now(), None, None);

        let result = handler.handle(cmd).await;
        assert!(matches!(
            result,
            Err(UpdateCourierLocationError::CourierNotFound(_))
        ));
    }
}
