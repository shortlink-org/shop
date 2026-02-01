//! Get Courier Handler
//!
//! Handles retrieving a single courier by ID.
//!
//! ## Flow
//! 1. Load courier from repository
//! 2. Get state from cache
//! 3. Optionally fetch current location from Geolocation Service
//! 4. Return courier with state

use std::sync::Arc;

use thiserror::Error;
use uuid::Uuid;

use crate::domain::ports::{
    CacheError, CachedCourierState, CourierCache, CourierRepository, QueryHandler, RepositoryError,
};
use crate::domain::model::courier::Courier;

use super::Query;

/// Errors that can occur during courier retrieval
#[derive(Debug, Error)]
pub enum GetCourierError {
    /// Courier not found
    #[error("Courier not found: {0}")]
    NotFound(Uuid),

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    /// Cache error
    #[error("Cache error: {0}")]
    CacheError(#[from] CacheError),
}

/// Response from get courier query
#[derive(Debug, Clone)]
pub struct Response {
    /// Courier profile from database
    pub courier: Courier,
    /// Cached state (status, load, rating)
    pub state: Option<CachedCourierState>,
}

/// Get Courier Handler
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
}

impl<R, C> QueryHandler<Query, Response> for Handler<R, C>
where
    R: CourierRepository + Send + Sync,
    C: CourierCache + Send + Sync,
{
    type Error = GetCourierError;

    /// Handle the GetCourier query
    async fn handle(&self, query: Query) -> Result<Response, Self::Error> {
        // 1. Load courier from repository
        let courier = self
            .repository
            .find_by_id(query.courier_id)
            .await?
            .ok_or(GetCourierError::NotFound(query.courier_id))?;

        // 2. Get state from cache
        let state = self.cache.get_state(query.courier_id).await.ok().flatten();

        // 3. TODO: Optionally fetch location from Geolocation Service
        // if query.include_location { ... }

        // 4. Return courier with state
        Ok(Response { courier, state })
    }
}

#[cfg(test)]
mod tests {
    // TODO: Add unit tests with mock repository and cache
}
