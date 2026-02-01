//! Archive Courier Handler
//!
//! Handles archiving a courier (soft delete).
//!
//! ## Flow
//! 1. Validate courier exists
//! 2. Check courier has no active deliveries
//! 3. Update status to ARCHIVED in cache
//! 4. Mark as archived in repository
//! 5. Remove from all pools
//! 6. Return archival result

use std::sync::Arc;

use chrono::{DateTime, Utc};
use thiserror::Error;
use uuid::Uuid;

use crate::boundary::ports::{
    CacheError, CommandHandlerWithResult, CourierCache, CourierRepository, RepositoryError,
};
use crate::domain::model::courier::CourierStatus;

use super::Command;

/// Errors that can occur during courier archival
#[derive(Debug, Error)]
pub enum ArchiveCourierError {
    /// Courier not found
    #[error("Courier not found: {0}")]
    NotFound(Uuid),

    /// Courier is already archived
    #[error("Courier is already archived: {0}")]
    AlreadyArchived(Uuid),

    /// Courier has active deliveries
    #[error("Cannot archive courier with active deliveries: {0}")]
    HasActiveDeliveries(Uuid),

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    /// Cache error
    #[error("Cache error: {0}")]
    CacheError(#[from] CacheError),
}

/// Response from archiving a courier
#[derive(Debug, Clone)]
pub struct Response {
    /// Courier ID
    pub courier_id: Uuid,
    /// New status
    pub status: CourierStatus,
    /// Archival timestamp
    pub archived_at: DateTime<Utc>,
}

/// Archive Courier Handler
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
    type Error = ArchiveCourierError;

    /// Handle the ArchiveCourier command
    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        // 1. Validate courier exists
        let courier = self
            .repository
            .find_by_id(cmd.courier_id)
            .await?
            .ok_or(ArchiveCourierError::NotFound(cmd.courier_id))?;

        // 2. Check current status from cache
        if let Ok(Some(state)) = self.cache.get_state(cmd.courier_id).await {
            if state.status == CourierStatus::Archived {
                return Err(ArchiveCourierError::AlreadyArchived(cmd.courier_id));
            }
            // Check if courier has active deliveries
            if state.current_load > 0 {
                return Err(ArchiveCourierError::HasActiveDeliveries(cmd.courier_id));
            }
        }

        // 3. Update status to ARCHIVED in cache
        self.cache
            .update_status(cmd.courier_id, CourierStatus::Archived)
            .await?;

        // 4. Mark as archived in repository
        self.repository.archive(cmd.courier_id).await?;

        // 5. Remove from all pools
        self.cache
            .remove_from_free_pool(cmd.courier_id, courier.work_zone())
            .await?;

        let archived_at = Utc::now();

        Ok(Response {
            courier_id: cmd.courier_id,
            status: CourierStatus::Archived,
            archived_at,
        })
    }
}

#[cfg(test)]
mod tests {
    // TODO: Add unit tests with mock repository and cache
}
