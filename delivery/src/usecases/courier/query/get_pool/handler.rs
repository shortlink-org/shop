//! Get Courier Pool Handler
//!
//! Handles retrieving couriers with filtering and pagination.
//!
//! ## Flow
//! 1. Build query from filters
//! 2. Get courier IDs from cache (for status-based filters)
//! 3. Load courier profiles from repository
//! 4. Optionally fetch current locations from Geolocation Service
//! 5. Return paginated results

use std::sync::Arc;

use thiserror::Error;
use uuid::Uuid;

use crate::domain::ports::{
    CacheError, CachedCourierState, CourierCache, CourierRepository, QueryHandler, RepositoryError,
};
use crate::domain::model::courier::{Courier, CourierStatus};

use super::Query;

/// Errors that can occur during courier pool retrieval
#[derive(Debug, Error)]
pub enum GetCourierPoolError {
    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    /// Cache error
    #[error("Cache error: {0}")]
    CacheError(#[from] CacheError),
}

/// Courier data with state from cache
#[derive(Debug, Clone)]
pub struct CourierWithState {
    /// Courier profile from database
    pub courier: Courier,
    /// Cached state (status, load, rating)
    pub state: Option<CachedCourierState>,
}

/// Response from get courier pool query
#[derive(Debug)]
pub struct Response {
    /// List of couriers with their states
    pub couriers: Vec<CourierWithState>,
    /// Total count of matching couriers
    pub total_count: usize,
}

/// Get Courier Pool Handler
pub struct Handler<R, C>
where
    R: CourierRepository,
    C: CourierCache,
{
    repository: Arc<R>,
    cache: Arc<C>,
}

impl<R, C> Handler<R, C>
where
    R: CourierRepository,
    C: CourierCache,
{
    /// Create a new handler instance
    pub fn new(repository: Arc<R>, cache: Arc<C>) -> Self {
        Self { repository, cache }
    }

    /// Get courier IDs based on filter
    async fn get_filtered_ids(&self, query: &Query) -> Result<Vec<Uuid>, GetCourierPoolError> {
        let filter = &query.filter;

        // If filtering by status = Free and zone is specified, use cache
        if filter.status == Some(CourierStatus::Free) {
            if let Some(ref zone) = filter.work_zone {
                return Ok(self.cache.get_free_couriers_in_zone(zone).await?);
            } else {
                return Ok(self.cache.get_all_free_couriers().await?);
            }
        }

        // Otherwise, query from repository
        if let Some(ref zone) = filter.work_zone {
            let couriers = self.repository.find_by_work_zone(zone).await?;
            return Ok(couriers.iter().map(|c| c.id().0).collect());
        }

        // For now, if no filters, we'd need a find_all method
        // This is a limitation - in production, add pagination and find_all
        Ok(vec![])
    }
}

impl<R, C> QueryHandler<Query, Response> for Handler<R, C>
where
    R: CourierRepository + Send + Sync,
    C: CourierCache + Send + Sync,
{
    type Error = GetCourierPoolError;

    /// Handle the GetCourierPool query
    async fn handle(&self, query: Query) -> Result<Response, Self::Error> {
        // Strategy depends on filters
        let courier_ids = self.get_filtered_ids(&query).await?;

        // Load couriers from repository and enrich with cache data
        let mut couriers_with_state = Vec::with_capacity(courier_ids.len());

        for id in courier_ids {
            if let Some(courier) = self.repository.find_by_id(id).await? {
                // Apply additional filters
                if let Some(ref transport) = query.filter.transport_type {
                    if courier.transport_type() != *transport {
                        continue;
                    }
                }

                // Get state from cache
                let state = self.cache.get_state(id).await.ok().flatten();

                // Apply availability filter
                if query.filter.available_only {
                    if let Some(ref s) = state {
                        if s.current_load >= s.max_load {
                            continue;
                        }
                    }
                }

                couriers_with_state.push(CourierWithState { courier, state });
            }
        }

        let total_count = couriers_with_state.len();

        Ok(Response {
            couriers: couriers_with_state,
            total_count,
        })
    }
}

#[cfg(test)]
mod tests {
    // TODO: Add unit tests with mock repository and cache
}
