//! Change Transport Type Handler
//!
//! Handles changing courier transport type.
//!
//! ## Flow
//! 1. Validate courier exists
//! 2. Check courier is not archived
//! 3. Check courier has no active deliveries (capacity change)
//! 4. Update transport type and recalculate max_load
//! 5. Update cache with new max_load
//! 6. Return update result

use std::sync::Arc;

use chrono::{DateTime, Utc};
use thiserror::Error;
use uuid::Uuid;

use crate::domain::ports::{
    CacheError, CommandHandlerWithResult, CourierCache, CourierRepository, RepositoryError,
};
use crate::domain::model::courier::CourierStatus;
use crate::domain::model::vo::TransportType;

use super::Command;

/// Errors that can occur during transport type change
#[derive(Debug, Error)]
pub enum ChangeTransportTypeError {
    /// Courier not found
    #[error("Courier not found: {0}")]
    NotFound(Uuid),

    /// Courier is archived
    #[error("Cannot update archived courier: {0}")]
    CourierArchived(Uuid),

    /// Courier has active deliveries
    #[error("Cannot change transport type while courier has active deliveries: {0}")]
    HasActiveDeliveries(Uuid),

    /// Same transport type
    #[error("Courier already has this transport type")]
    SameTransportType,

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    /// Cache error
    #[error("Cache error: {0}")]
    CacheError(#[from] CacheError),
}

/// Response from changing transport type
#[derive(Debug, Clone)]
pub struct Response {
    /// Courier ID
    pub courier_id: Uuid,
    /// New transport type
    pub transport_type: TransportType,
    /// New max load (recalculated based on transport)
    pub max_load: u32,
    /// Update timestamp
    pub updated_at: DateTime<Utc>,
}

/// Change Transport Type Handler
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
    type Error = ChangeTransportTypeError;

    /// Handle the ChangeTransportType command
    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        // 1. Validate courier exists
        let mut courier = self
            .repository
            .find_by_id(cmd.courier_id)
            .await?
            .ok_or(ChangeTransportTypeError::NotFound(cmd.courier_id))?;

        // Check if transport type is the same
        if courier.transport_type() == cmd.transport_type {
            return Err(ChangeTransportTypeError::SameTransportType);
        }

        // 2. Check courier is not archived and has no active deliveries
        if let Ok(Some(state)) = self.cache.get_state(cmd.courier_id).await {
            if state.status == CourierStatus::Archived {
                return Err(ChangeTransportTypeError::CourierArchived(cmd.courier_id));
            }
            // Cannot change transport type while having active deliveries
            // because max_load will change
            if state.current_load > 0 {
                return Err(ChangeTransportTypeError::HasActiveDeliveries(cmd.courier_id));
            }
        }

        // 3. Update transport type (this recalculates max_load internally)
        courier.change_transport_type(cmd.transport_type);
        let new_max_load = courier.max_load();

        // 4. Save to repository
        self.repository.save(&courier).await?;

        // 5. Update max_load in cache
        self.cache
            .update_max_load(cmd.courier_id, new_max_load)
            .await?;

        let updated_at = Utc::now();

        Ok(Response {
            courier_id: cmd.courier_id,
            transport_type: cmd.transport_type,
            max_load: new_max_load,
            updated_at,
        })
    }
}

#[cfg(test)]
mod tests {
    // TODO: Add unit tests with mock repository and cache
}
