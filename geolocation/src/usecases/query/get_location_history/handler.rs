//! Get Location History Handler
//!
//! Fetches location history for a courier within a time range.

use std::sync::Arc;

use async_trait::async_trait;
use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::LocationHistoryEntry;
use crate::domain::ports::{LocationRepository, QueryHandler, RepositoryError};

use super::Query;

/// Errors that can occur during history fetch
#[derive(Debug, Error)]
pub enum GetLocationHistoryError {
    #[error("Invalid time range")]
    InvalidTimeRange,

    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),
}

/// Response containing location history
#[derive(Debug, Clone)]
pub struct Response {
    /// Courier ID
    pub courier_id: Uuid,
    /// History entries
    pub entries: Vec<LocationHistoryEntry>,
    /// Total count of entries returned
    pub count: usize,
}

/// Get Location History Handler
pub struct Handler<R>
where
    R: LocationRepository,
{
    location_repo: Arc<R>,
}

impl<R> Handler<R>
where
    R: LocationRepository,
{
    /// Create a new handler
    pub fn new(location_repo: Arc<R>) -> Self {
        Self { location_repo }
    }
}

#[async_trait]
impl<R> QueryHandler<Query, Response> for Handler<R>
where
    R: LocationRepository + Send + Sync,
{
    type Error = GetLocationHistoryError;

    async fn handle(&self, query: Query) -> Result<Response, Self::Error> {
        let limit = query.effective_limit();

        let entries = self
            .location_repo
            .get_location_history(query.courier_id, &query.time_range, Some(limit))
            .await?;

        let count = entries.len();

        Ok(Response {
            courier_id: query.courier_id,
            entries,
            count,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::{CourierLocation, Location, TimeRange};
    use async_trait::async_trait;
    use chrono::{Duration, Utc};
    use std::sync::Mutex;

    // === Mock Repository ===

    struct MockLocationRepository {
        history: Mutex<Vec<LocationHistoryEntry>>,
    }

    impl MockLocationRepository {
        fn new() -> Self {
            Self {
                history: Mutex::new(Vec::new()),
            }
        }

        fn add_history_entry(&self, entry: LocationHistoryEntry) {
            let mut history = self.history.lock().unwrap();
            history.push(entry);
        }
    }

    #[async_trait]
    impl LocationRepository for MockLocationRepository {
        async fn save_current_location(&self, _location: &CourierLocation) -> Result<(), RepositoryError> {
            Ok(())
        }

        async fn get_current_location(&self, _courier_id: Uuid) -> Result<Option<CourierLocation>, RepositoryError> {
            Ok(None)
        }

        async fn get_current_locations(&self, _courier_ids: &[Uuid]) -> Result<Vec<CourierLocation>, RepositoryError> {
            Ok(vec![])
        }

        async fn save_history_entry(&self, entry: &LocationHistoryEntry) -> Result<(), RepositoryError> {
            let mut history = self.history.lock().unwrap();
            history.push(entry.clone());
            Ok(())
        }

        async fn get_location_history(
            &self,
            courier_id: Uuid,
            time_range: &TimeRange,
            limit: Option<u32>,
        ) -> Result<Vec<LocationHistoryEntry>, RepositoryError> {
            let history = self.history.lock().unwrap();
            let mut entries: Vec<_> = history
                .iter()
                .filter(|e| e.courier_id() == courier_id && time_range.contains(e.recorded_at()))
                .cloned()
                .collect();

            entries.sort_by(|a, b| b.recorded_at().cmp(&a.recorded_at())); // Most recent first

            if let Some(limit) = limit {
                entries.truncate(limit as usize);
            }

            Ok(entries)
        }

        async fn delete_history_before(&self, _before: chrono::DateTime<Utc>) -> Result<u64, RepositoryError> {
            Ok(0)
        }
    }

    // === Tests ===

    #[tokio::test]
    async fn test_get_location_history() {
        let repo = Arc::new(MockLocationRepository::new());

        let courier_id = Uuid::new_v4();
        let now = Utc::now();

        // Add some history entries
        for i in 0..5 {
            let location = Location::from_stored(
                52.52 + (i as f64 * 0.001),
                13.405,
                15.0,
                now - Duration::minutes(i * 10),
                None,
                None,
            )
            .unwrap();
            let entry = LocationHistoryEntry::new(courier_id, location);
            repo.add_history_entry(entry);
        }

        let handler = Handler::new(repo);
        let query = Query::new(
            courier_id,
            now - Duration::hours(1),
            now + Duration::minutes(1),
            Some(10),
        )
        .unwrap();

        let result = handler.handle(query).await;
        assert!(result.is_ok());

        let response = result.unwrap();
        assert_eq!(response.courier_id, courier_id);
        assert_eq!(response.count, 5);
    }
}
