//! Accept Order Handler
//!
//! Handles accepting an order from OMS for delivery.
//!
//! ## Flow
//! 1. Validate order data
//! 2. Create Package aggregate
//! 3. Transition to IN_POOL status
//! 4. Save to repository
//! 5. Generate PackageAcceptedEvent

use std::sync::Arc;

use thiserror::Error;
use uuid::Uuid;

use crate::domain::ports::{CommandHandlerWithResult, CourierRepository, RepositoryError};

use super::Command;

/// Errors that can occur during order acceptance
#[derive(Debug, Error)]
pub enum AcceptOrderError {
    /// Order already exists
    #[error("Order already accepted: {0}")]
    OrderAlreadyExists(String),

    /// Invalid order data
    #[error("Invalid order data: {0}")]
    InvalidData(String),

    /// Invalid location
    #[error("Invalid location: {0}")]
    InvalidLocation(String),

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),
}

/// Response from accepting an order
#[derive(Debug, Clone)]
pub struct Response {
    /// The created package ID
    pub package_id: Uuid,
    /// The order ID from OMS
    pub order_id: Uuid,
}

/// Accept Order Handler
pub struct Handler<R>
where
    R: CourierRepository,
{
    #[allow(dead_code)]
    repository: Arc<R>,
    // TODO: Add PackageRepository when implemented
}

impl<R> Handler<R>
where
    R: CourierRepository,
{
    /// Create a new handler instance
    pub fn new(repository: Arc<R>) -> Self {
        Self { repository }
    }
}

impl<R> CommandHandlerWithResult<Command, Response> for Handler<R>
where
    R: CourierRepository + Send + Sync,
{
    type Error = AcceptOrderError;

    /// Handle the AcceptOrder command
    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        // 1. Validate order data
        if cmd.weight_kg <= 0.0 {
            return Err(AcceptOrderError::InvalidData(
                "Package weight must be positive".to_string(),
            ));
        }

        if cmd.priority == 0 || cmd.priority > 5 {
            return Err(AcceptOrderError::InvalidData(
                "Priority must be between 1 and 5".to_string(),
            ));
        }

        // 2. Locations are already validated at construction time
        // (Location::new() returns Result and validates coordinates)

        // 3. Create Package aggregate
        // TODO: Implement Package aggregate creation
        let package_id = Uuid::new_v4();

        // 4. Transition to IN_POOL status
        // TODO: package.transition_to(PackageStatus::InPool)?;

        // 5. Save to repository
        // TODO: self.package_repository.save(&package).await?;

        // 6. Generate PackageAcceptedEvent
        // TODO: Publish event to message broker

        Ok(Response {
            package_id,
            order_id: cmd.order_id,
        })
    }
}

#[cfg(test)]
mod tests {
    // TODO: Add unit tests with mock repository
}
