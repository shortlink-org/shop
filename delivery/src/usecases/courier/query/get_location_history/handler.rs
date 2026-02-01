//! Get Location History Handler
//!
//! Handles retrieving a courier's location history.
//!
//! ## Flow
//! 1. Query repository for location history within time range
//! 2. Return paginated results

use std::sync::Arc;

use thiserror::Error;

use crate::domain::model::LocationHistoryEntry;
use crate::domain::ports::{LocationRepository, LocationRepositoryError, QueryHandler};

use super::Query;

/// Errors that can occur during history retrieval
#[derive(Debug, Error)]
pub enum GetLocationHistoryError {
    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] LocationRepositoryError),
}

/// Response from get location history query
#[derive(Debug, Clone)]
pub struct Response {
    /// Location history entries (ordered by timestamp)
    pub entries: Vec<LocationHistoryEntry>,
    /// Total count of entries in the time range
    pub total_count: u64,
    /// Whether there are more entries (for pagination)
    pub has_more: bool,
}

/// Get Location History Handler
pub struct Handler<LR>
where
    LR: LocationRepository,
{
    location_repository: Arc<LR>,
}

impl<LR> Handler<LR>
where
    LR: LocationRepository,
{
    /// Create a new handler instance
    pub fn new(location_repository: Arc<LR>) -> Self {
        Self { location_repository }
    }
}

impl<LR> QueryHandler<Query, Response> for Handler<LR>
where
    LR: LocationRepository + Send + Sync,
{
    type Error = GetLocationHistoryError;

    /// Handle the GetLocationHistory query
    async fn handle(&self, query: Query) -> Result<Response, Self::Error> {
        // Get total count
        let total_count = self
            .location_repository
            .count_history(query.courier_id, query.time_range)
            .await?;

        // Get entries (with or without pagination)
        let entries = if let (Some(limit), Some(offset)) = (query.limit, query.offset) {
            self.location_repository
                .get_history_paginated(query.courier_id, query.time_range, limit, offset)
                .await?
        } else {
            self.location_repository
                .get_history(query.courier_id, query.time_range)
                .await?
        };

        // Check if there are more entries
        let has_more = if let (Some(limit), Some(offset)) = (query.limit, query.offset) {
            (offset + limit) < total_count as u32
        } else {
            false
        };

        Ok(Response {
            entries,
            total_count,
            has_more,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::courier_location::TimeRange;
    use crate::domain::model::vo::location::Location;
    use async_trait::async_trait;
    use chrono::Utc;
    use std::sync::Mutex;
    use uuid::Uuid;

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
    async fn test_get_location_history() {
        let courier_id = Uuid::new_v4();
        let repo = Arc::new(MockLocationRepository::new());

        // Add some history entries
        let now = Utc::now();
        for i in 0..5 {
            let entry = LocationHistoryEntry::new(
                courier_id,
                Location::new(52.52 + i as f64 * 0.001, 13.405, 10.0).unwrap(),
                now - chrono::Duration::minutes(i as i64),
                Some(35.0),
                None,
            );
            repo.add(entry);
        }

        let handler = Handler::new(repo);
        let time_range = TimeRange::new(now - chrono::Duration::hours(1), now).unwrap();
        let query = Query::new(courier_id, time_range);

        let result = handler.handle(query).await;
        assert!(result.is_ok());

        let response = result.unwrap();
        assert_eq!(response.entries.len(), 5);
        assert_eq!(response.total_count, 5);
        assert!(!response.has_more);
    }

    #[tokio::test]
    async fn test_get_location_history_paginated() {
        let courier_id = Uuid::new_v4();
        let repo = Arc::new(MockLocationRepository::new());

        // Add 10 history entries
        let now = Utc::now();
        for i in 0..10 {
            let entry = LocationHistoryEntry::new(
                courier_id,
                Location::new(52.52 + i as f64 * 0.001, 13.405, 10.0).unwrap(),
                now - chrono::Duration::minutes(i as i64),
                None,
                None,
            );
            repo.add(entry);
        }

        let handler = Handler::new(repo);
        let time_range = TimeRange::new(now - chrono::Duration::hours(1), now).unwrap();

        // Get first page
        let query = Query::paginated(courier_id, time_range, 5, 0);
        let result = handler.handle(query).await.unwrap();
        assert_eq!(result.entries.len(), 5);
        assert_eq!(result.total_count, 10);
        assert!(result.has_more);

        // Get second page
        let query = Query::paginated(courier_id, time_range, 5, 5);
        let result = handler.handle(query).await.unwrap();
        assert_eq!(result.entries.len(), 5);
        assert!(!result.has_more);
    }
}
