//! Update Work Schedule Handler
//!
//! Handles updating courier work schedule.
//!
//! ## Flow
//! 1. Validate courier exists
//! 2. Check courier is not archived
//! 3. If work zone changes, update cache pools
//! 4. Update courier in repository
//! 5. Return update result

use std::sync::Arc;

use chrono::{DateTime, Utc};
use thiserror::Error;
use uuid::Uuid;

use crate::domain::ports::{
    CacheError, CommandHandlerWithResult, CourierCache, CourierRepository, RepositoryError,
};
use crate::domain::model::courier::{CourierStatus, WorkHours};

use super::Command;

/// Errors that can occur during work schedule update
#[derive(Debug, Error)]
pub enum UpdateWorkScheduleError {
    /// Courier not found
    #[error("Courier not found: {0}")]
    NotFound(Uuid),

    /// Courier is archived
    #[error("Cannot update archived courier: {0}")]
    CourierArchived(Uuid),

    /// No fields to update
    #[error("No fields to update")]
    NoFieldsToUpdate,

    /// Invalid work hours
    #[error("Invalid work hours: {0}")]
    InvalidWorkHours(String),

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    /// Cache error
    #[error("Cache error: {0}")]
    CacheError(#[from] CacheError),
}

/// Response from updating work schedule
#[derive(Debug, Clone)]
pub struct Response {
    /// Courier ID
    pub courier_id: Uuid,
    /// Current work hours
    pub work_hours: WorkHours,
    /// Current work zone
    pub work_zone: String,
    /// Current max distance
    pub max_distance_km: f64,
    /// Update timestamp
    pub updated_at: DateTime<Utc>,
}

/// Update Work Schedule Handler
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
    type Error = UpdateWorkScheduleError;

    /// Handle the UpdateWorkSchedule command
    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        // Check if there's anything to update
        if cmd.work_hours.is_none() && cmd.work_zone.is_none() && cmd.max_distance_km.is_none() {
            return Err(UpdateWorkScheduleError::NoFieldsToUpdate);
        }

        // 1. Validate courier exists
        let mut courier = self
            .repository
            .find_by_id(cmd.courier_id)
            .await?
            .ok_or(UpdateWorkScheduleError::NotFound(cmd.courier_id))?;

        // 2. Check courier is not archived
        let current_status = if let Ok(Some(state)) = self.cache.get_state(cmd.courier_id).await {
            if state.status == CourierStatus::Archived {
                return Err(UpdateWorkScheduleError::CourierArchived(cmd.courier_id));
            }
            Some(state.status)
        } else {
            None
        };

        let old_zone = courier.work_zone().to_string();

        // 3. Update courier fields
        if let Some(work_hours) = cmd.work_hours {
            courier.update_work_hours(work_hours);
        }
        if let Some(ref work_zone) = cmd.work_zone {
            courier.update_work_zone(work_zone.clone());
        }
        if let Some(max_distance_km) = cmd.max_distance_km {
            courier.update_max_distance(max_distance_km);
        }

        // 4. If work zone changed and courier is FREE, update cache pools
        if let Some(ref new_zone) = cmd.work_zone {
            if new_zone != &old_zone {
                if let Some(CourierStatus::Free) = current_status {
                    // Remove from old zone pool
                    self.cache
                        .remove_from_free_pool(cmd.courier_id, &old_zone)
                        .await?;
                    // Add to new zone pool
                    self.cache
                        .add_to_free_pool(cmd.courier_id, new_zone)
                        .await?;
                }
            }
        }

        // 5. Save to repository
        self.repository.save(&courier).await?;

        let updated_at = Utc::now();

        Ok(Response {
            courier_id: cmd.courier_id,
            work_hours: courier.work_hours().clone(),
            work_zone: courier.work_zone().to_string(),
            max_distance_km: courier.max_distance_km(),
            updated_at,
        })
    }
}

#[cfg(test)]
mod tests {
    // TODO: Add unit tests with mock repository and cache
}
