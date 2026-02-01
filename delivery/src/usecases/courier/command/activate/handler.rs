//! Activate Courier Handler
//!
//! Handles activating a courier (setting status to FREE).
//!
//! ## Flow
//! 1. Validate courier exists
//! 2. Check current status (cannot activate archived courier)
//! 3. Update status to FREE in cache
//! 4. Return activation result

use std::sync::Arc;

use chrono::{DateTime, Utc};
use thiserror::Error;
use uuid::Uuid;

use crate::domain::ports::{
    CacheError, CommandHandlerWithResult, CourierCache, CourierRepository, RepositoryError,
};
use crate::domain::model::courier::CourierStatus;

use super::Command;

/// Errors that can occur during courier activation
#[derive(Debug, Error)]
pub enum ActivateCourierError {
    /// Courier not found
    #[error("Courier not found: {0}")]
    NotFound(Uuid),

    /// Courier is archived and cannot be activated
    #[error("Cannot activate archived courier: {0}")]
    CourierArchived(Uuid),

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    /// Cache error
    #[error("Cache error: {0}")]
    CacheError(#[from] CacheError),
}

/// Response from activating a courier
#[derive(Debug, Clone)]
pub struct Response {
    /// Courier ID
    pub courier_id: Uuid,
    /// New status
    pub status: CourierStatus,
    /// Activation timestamp
    pub activated_at: DateTime<Utc>,
}

/// Activate Courier Handler
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

impl<R, C> CommandHandlerWithResult<Command, Response> for Handler<R, C>
where
    R: CourierRepository + Send + Sync,
    C: CourierCache + Send + Sync,
{
    type Error = ActivateCourierError;

    /// Handle the ActivateCourier command
    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        // 1. Validate courier exists
        let courier = self
            .repository
            .find_by_id(cmd.courier_id)
            .await?
            .ok_or(ActivateCourierError::NotFound(cmd.courier_id))?;

        // 2. Check current status from cache
        if let Ok(Some(state)) = self.cache.get_state(cmd.courier_id).await {
            if state.status == CourierStatus::Archived {
                return Err(ActivateCourierError::CourierArchived(cmd.courier_id));
            }
        }

        // 3. Update status to FREE in cache
        self.cache
            .update_status(cmd.courier_id, CourierStatus::Free)
            .await?;

        // 4. Add to free couriers set for the zone
        self.cache
            .add_to_free_pool(cmd.courier_id, courier.work_zone())
            .await?;

        let activated_at = Utc::now();

        Ok(Response {
            courier_id: cmd.courier_id,
            status: CourierStatus::Free,
            activated_at,
        })
    }
}

#[cfg(test)]
mod tests {
    // TODO: Add unit tests with mock repository and cache
}
