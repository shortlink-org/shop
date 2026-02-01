//! Deactivate Courier Handler
//!
//! Handles deactivating a courier (setting status to UNAVAILABLE).
//!
//! ## Flow
//! 1. Validate courier exists
//! 2. Check current status (cannot deactivate archived courier)
//! 3. Update status to UNAVAILABLE in cache
//! 4. Remove from free pool
//! 5. Return deactivation result

use std::sync::Arc;

use chrono::{DateTime, Utc};
use thiserror::Error;
use uuid::Uuid;

use crate::domain::ports::{
    CacheError, CommandHandlerWithResult, CourierCache, CourierRepository, RepositoryError,
};
use crate::domain::model::courier::CourierStatus;

use super::Command;

/// Errors that can occur during courier deactivation
#[derive(Debug, Error)]
pub enum DeactivateCourierError {
    /// Courier not found
    #[error("Courier not found: {0}")]
    NotFound(Uuid),

    /// Courier is archived and cannot be deactivated
    #[error("Cannot deactivate archived courier: {0}")]
    CourierArchived(Uuid),

    /// Courier has active deliveries
    #[error("Courier has active deliveries: {0}")]
    HasActiveDeliveries(Uuid),

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    /// Cache error
    #[error("Cache error: {0}")]
    CacheError(#[from] CacheError),
}

/// Response from deactivating a courier
#[derive(Debug, Clone)]
pub struct Response {
    /// Courier ID
    pub courier_id: Uuid,
    /// New status
    pub status: CourierStatus,
    /// Deactivation timestamp
    pub deactivated_at: DateTime<Utc>,
}

/// Deactivate Courier Handler
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
    type Error = DeactivateCourierError;

    /// Handle the DeactivateCourier command
    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        // 1. Validate courier exists
        let courier = self
            .repository
            .find_by_id(cmd.courier_id)
            .await?
            .ok_or(DeactivateCourierError::NotFound(cmd.courier_id))?;

        // 2. Check current status from cache
        if let Ok(Some(state)) = self.cache.get_state(cmd.courier_id).await {
            if state.status == CourierStatus::Archived {
                return Err(DeactivateCourierError::CourierArchived(cmd.courier_id));
            }
            // Check if courier has active deliveries
            if state.current_load > 0 {
                return Err(DeactivateCourierError::HasActiveDeliveries(cmd.courier_id));
            }
        }

        // 3. Update status to UNAVAILABLE in cache
        self.cache
            .update_status(cmd.courier_id, CourierStatus::Unavailable)
            .await?;

        // 4. Remove from free pool
        self.cache
            .remove_from_free_pool(cmd.courier_id, courier.work_zone())
            .await?;

        let deactivated_at = Utc::now();

        Ok(Response {
            courier_id: cmd.courier_id,
            status: CourierStatus::Unavailable,
            deactivated_at,
        })
    }
}

#[cfg(test)]
mod tests {
    // TODO: Add unit tests with mock repository and cache
}
