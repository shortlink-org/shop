//! Assign Order Handler
//!
//! Handles assigning a package to a courier.
//!
//! ## Flow
//! 1. Load package from repository
//! 2. If auto-assign: use DispatchService to find nearest courier
//! 3. If manual: validate assignment using AssignmentValidationService
//! 4. Update package status to ASSIGNED
//! 5. Update courier status to BUSY
//! 6. Generate PackageAssignedEvent
//! 7. Send push notification

use std::sync::Arc;

use thiserror::Error;
use uuid::Uuid;

use crate::domain::ports::{
    CommandHandlerWithResult, CourierCache, CourierRepository, RepositoryError,
};
// Note: DispatchService will be used when PackageRepository is implemented
// use crate::domain::services::dispatch::DispatchService;

use super::command::AssignmentMode;
use super::Command;

/// Errors that can occur during order assignment
#[derive(Debug, Error)]
pub enum AssignOrderError {
    /// Package not found
    #[error("Package not found: {0}")]
    PackageNotFound(String),

    /// Courier not found
    #[error("Courier not found: {0}")]
    CourierNotFound(String),

    /// No available courier
    #[error("No available courier for package: {0}")]
    NoAvailableCourier(String),

    /// Courier ID required for manual assignment
    #[error("Courier ID required for manual assignment")]
    CourierIdRequired,

    /// Assignment validation failed
    #[error("Assignment validation failed: {0}")]
    ValidationFailed(String),

    /// Invalid package status for assignment
    #[error("Invalid package status for assignment: {0}")]
    InvalidPackageStatus(String),

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),
}

/// Response from assigning an order
#[derive(Debug, Clone)]
pub struct Response {
    /// The package ID
    pub package_id: Uuid,
    /// The assigned courier ID
    pub courier_id: Uuid,
    /// Estimated delivery time in minutes
    pub estimated_delivery_minutes: u32,
}

/// Assign Order Handler
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
    type Error = AssignOrderError;

    /// Handle the AssignOrder command
    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        // 1. Load package from repository
        // TODO: Implement PackageRepository
        // let package = self.package_repository.find_by_id(cmd.package_id).await?
        //     .ok_or_else(|| AssignOrderError::PackageNotFound(cmd.package_id.to_string()))?;

        // 2. Find courier based on assignment mode
        let courier_id = match cmd.mode {
            AssignmentMode::Auto => {
                // Use DispatchService to find the best courier
                // TODO: Get package zone and location
                let zone = "default"; // Placeholder
                let free_courier_ids = self.cache.get_free_couriers_in_zone(zone).await.map_err(|e| {
                    AssignOrderError::RepositoryError(RepositoryError::QueryError(e.to_string()))
                })?;

                if free_courier_ids.is_empty() {
                    return Err(AssignOrderError::NoAvailableCourier(
                        cmd.package_id.to_string(),
                    ));
                }

                // Load courier details
                let mut couriers = Vec::new();
                for id in free_courier_ids {
                    if let Some(courier) = self.repository.find_by_id(id).await? {
                        couriers.push(courier);
                    }
                }

                // TODO: Get package location and create PackageForDispatch for proper dispatch
                // For now, just pick the first available courier
                let best_courier = couriers.first().ok_or_else(|| {
                    AssignOrderError::NoAvailableCourier(cmd.package_id.to_string())
                })?;

                best_courier.id().0
            }
            AssignmentMode::Manual => {
                let courier_id = cmd
                    .courier_id
                    .ok_or(AssignOrderError::CourierIdRequired)?;

                // Validate courier exists and is available
                let _courier =
                    self.repository
                        .find_by_id(courier_id)
                        .await?
                        .ok_or_else(|| {
                            AssignOrderError::CourierNotFound(courier_id.to_string())
                        })?;

                // TODO: Validate using AssignmentValidationService
                // self.validation_service.validate_assignment(&package, &courier)?;

                courier_id
            }
        };

        // 3. Update package status to ASSIGNED
        // TODO: package.transition_to(PackageStatus::Assigned)?;

        // 4. Update courier status to BUSY
        // TODO: self.cache.update_status(courier_id, CourierStatus::Busy).await?;

        // 5. Save changes
        // TODO: self.package_repository.save(&package).await?;

        // 6. Generate PackageAssignedEvent
        // TODO: Publish event

        // 7. Send push notification
        // TODO: self.notification_service.send_assignment_notification(courier_id, package_id).await?;

        Ok(Response {
            package_id: cmd.package_id,
            courier_id,
            estimated_delivery_minutes: 30, // TODO: Calculate actual estimate
        })
    }
}

#[cfg(test)]
mod tests {
    // TODO: Add unit tests with mock repository and cache
}
