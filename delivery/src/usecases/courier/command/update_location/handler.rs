//! Update Courier Location Handler
//!
//! Handles updating a courier's GPS location in real-time.
//!
//! ## Flow
//! 1. Validate location data
//! 2. Save location to Geolocation Service
//! 3. Generate CourierLocationUpdatedEvent

use std::sync::Arc;

use thiserror::Error;

use crate::domain::ports::{CommandHandler, CourierCache, CourierRepository, RepositoryError};

use super::Command;

/// Errors that can occur during courier location update
#[derive(Debug, Error)]
pub enum UpdateCourierLocationError {
    /// Courier not found
    #[error("Courier not found: {0}")]
    CourierNotFound(String),

    /// Invalid location data
    #[error("Invalid location data: {0}")]
    InvalidLocation(String),

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    /// Geolocation service error
    #[error("Geolocation service error: {0}")]
    GeolocationError(String),
}

/// Update Courier Location Handler
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
    type Error = UpdateCourierLocationError;

    /// Handle the UpdateCourierLocation command
    async fn handle(&self, cmd: Command) -> Result<(), Self::Error> {
        // 1. Validate courier exists
        let _courier = self
            .repository
            .find_by_id(cmd.courier_id)
            .await?
            .ok_or_else(|| {
                UpdateCourierLocationError::CourierNotFound(cmd.courier_id.to_string())
            })?;

        // 2. Location is already validated at construction time
        // (Location::new() returns Result and validates coordinates)

        // 3. Save location to Geolocation Service
        // TODO: Integrate with Geolocation gRPC service
        // geolocation_client.save_location(SaveLocationRequest {
        //     courier_id: cmd.courier_id,
        //     location: cmd.location,
        //     timestamp_ms: cmd.timestamp_ms,
        // }).await?;

        // 4. Generate CourierLocationUpdatedEvent
        // TODO: Publish event to message broker
        // event_publisher.publish(CourierLocationUpdatedEvent {
        //     courier_id: cmd.courier_id,
        //     location: cmd.location,
        //     timestamp_ms: cmd.timestamp_ms,
        // }).await?;

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    // TODO: Add unit tests with mock repository and geolocation service
}
