//! Deliver Order Handler
//!
//! Handles confirming delivery by courier.
//!
//! ## Flow
//! 1. Load package from repository
//! 2. Validate courier is assigned to package
//! 3. Update package status (DELIVERED or NOT_DELIVERED)
//! 4. Update courier load
//! 5. Save courier location
//! 6. Generate PackageDeliveredEvent or PackageNotDeliveredEvent
//! 7. Notify OMS

use std::sync::Arc;

use thiserror::Error;

use crate::boundary::ports::{CommandHandler, CourierCache, CourierRepository, RepositoryError};

use super::command::DeliveryResult;
use super::Command;

/// Errors that can occur during delivery confirmation
#[derive(Debug, Error)]
pub enum DeliverOrderError {
    /// Package not found
    #[error("Package not found: {0}")]
    PackageNotFound(String),

    /// Courier not found
    #[error("Courier not found: {0}")]
    CourierNotFound(String),

    /// Courier not assigned to package
    #[error("Courier {0} is not assigned to package {1}")]
    CourierNotAssigned(String, String),

    /// Invalid package status
    #[error("Invalid package status for delivery: {0}")]
    InvalidPackageStatus(String),

    /// Missing reason for failed delivery
    #[error("Reason required for failed delivery")]
    MissingNotDeliveredReason,

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),
}

/// Deliver Order Handler
pub struct Handler<R, C>
where
    R: CourierRepository,
    C: CourierCache,
{
    repository: Arc<R>,
    #[allow(dead_code)]
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

impl<R, C> CommandHandler<Command> for Handler<R, C>
where
    R: CourierRepository + Send + Sync,
    C: CourierCache + Send + Sync,
{
    type Error = DeliverOrderError;

    /// Handle the DeliverOrder command
    async fn handle(&self, cmd: Command) -> Result<(), Self::Error> {
        // 1. Validate courier exists
        let _courier = self
            .repository
            .find_by_id(cmd.courier_id)
            .await?
            .ok_or_else(|| DeliverOrderError::CourierNotFound(cmd.courier_id.to_string()))?;

        // 2. Load package from repository
        // TODO: Implement PackageRepository
        // let package = self.package_repository.find_by_id(cmd.package_id).await?
        //     .ok_or_else(|| DeliverOrderError::PackageNotFound(cmd.package_id.to_string()))?;

        // 3. Validate courier is assigned to package
        // TODO: if package.assigned_courier_id != Some(cmd.courier_id) {
        //     return Err(DeliverOrderError::CourierNotAssigned(...));
        // }

        // 4. Validate not_delivered_reason if result is NotDelivered
        if cmd.result == DeliveryResult::NotDelivered && cmd.not_delivered_reason.is_none() {
            return Err(DeliverOrderError::MissingNotDeliveredReason);
        }

        // 5. Update package status
        match cmd.result {
            DeliveryResult::Delivered => {
                // TODO: package.transition_to(PackageStatus::Delivered)?;

                // Update courier stats
                // TODO: self.cache.increment_successful_deliveries(cmd.courier_id).await?;
            }
            DeliveryResult::NotDelivered => {
                // TODO: package.transition_to(PackageStatus::NotDelivered)?;

                // Update courier stats
                // TODO: self.cache.increment_failed_deliveries(cmd.courier_id).await?;
            }
        }

        // 6. Update courier load (decrease by 1)
        // TODO: self.cache.decrement_load(cmd.courier_id).await?;

        // 7. Save courier location
        // TODO: geolocation_client.save_location(cmd.confirmation_location).await?;

        // 8. Save package changes
        // TODO: self.package_repository.save(&package).await?;

        // 9. Generate and publish event
        // TODO: match cmd.result {
        //     DeliveryResult::Delivered => event_publisher.publish(PackageDeliveredEvent { ... }),
        //     DeliveryResult::NotDelivered => event_publisher.publish(PackageNotDeliveredEvent { ... }),
        // }

        // 10. Notify OMS
        // TODO: oms_client.notify_delivery_result(cmd.package_id, cmd.result).await?;

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    // TODO: Add unit tests with mock repositories
}
