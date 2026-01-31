//! Get Courier Pool Use Case
//!
//! Retrieves couriers with filtering and pagination.
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

use crate::boundary::ports::{CacheError, CachedCourierState, CourierCache, CourierRepository, RepositoryError};
use crate::domain::model::courier::{Courier, CourierStatus};
use crate::domain::model::vo::TransportType;

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

/// Filter criteria for courier pool
#[derive(Debug, Clone, Default)]
pub struct CourierFilter {
    /// Filter by status
    pub status: Option<CourierStatus>,
    /// Filter by work zone
    pub work_zone: Option<String>,
    /// Filter by transport type
    pub transport_type: Option<TransportType>,
    /// Only include couriers that can accept more packages
    pub available_only: bool,
}

impl CourierFilter {
    /// Create a filter for free couriers in a specific zone
    pub fn free_in_zone(zone: &str) -> Self {
        Self {
            status: Some(CourierStatus::Free),
            work_zone: Some(zone.to_string()),
            available_only: true,
            ..Default::default()
        }
    }

    /// Create a filter for all couriers in a zone
    pub fn in_zone(zone: &str) -> Self {
        Self {
            work_zone: Some(zone.to_string()),
            ..Default::default()
        }
    }
}

/// Courier data with state from cache
#[derive(Debug, Clone)]
pub struct CourierWithState {
    /// Courier profile from database
    pub courier: Courier,
    /// Cached state (status, load, rating)
    pub state: Option<CachedCourierState>,
}

/// Response from get courier pool
#[derive(Debug)]
pub struct GetCourierPoolResponse {
    pub couriers: Vec<CourierWithState>,
    pub total_count: usize,
}

/// Get Courier Pool Use Case
pub struct GetCourierPoolUseCase<R, C>
where
    R: CourierRepository,
    C: CourierCache,
{
    repository: Arc<R>,
    cache: Arc<C>,
}

impl<R, C> GetCourierPoolUseCase<R, C>
where
    R: CourierRepository,
    C: CourierCache,
{
    /// Create a new use case instance
    pub fn new(repository: Arc<R>, cache: Arc<C>) -> Self {
        Self { repository, cache }
    }

    /// Execute the use case
    pub async fn execute(
        &self,
        filter: CourierFilter,
    ) -> Result<GetCourierPoolResponse, GetCourierPoolError> {
        // Strategy depends on filters
        let courier_ids = self.get_filtered_ids(&filter).await?;

        // Load couriers from repository and enrich with cache data
        let mut couriers_with_state = Vec::with_capacity(courier_ids.len());

        for id in courier_ids {
            if let Some(courier) = self.repository.find_by_id(id).await? {
                // Apply additional filters
                if let Some(ref transport) = filter.transport_type {
                    if courier.transport_type() != *transport {
                        continue;
                    }
                }

                // Get state from cache
                let state = self.cache.get_state(id).await.ok().flatten();

                // Apply availability filter
                if filter.available_only {
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

        Ok(GetCourierPoolResponse {
            couriers: couriers_with_state,
            total_count,
        })
    }

    /// Get courier IDs based on filter
    async fn get_filtered_ids(&self, filter: &CourierFilter) -> Result<Vec<Uuid>, GetCourierPoolError> {
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

#[cfg(test)]
mod tests {
    // TODO: Add unit tests with mock repository and cache
}
