//! Register Courier Handler
//!
//! Handles the registration of a new courier.
//!
//! ## Flow
//! 1. Validate courier data
//! 2. Check for duplicate email/phone
//! 3. Create Courier aggregate
//! 4. Save to PostgreSQL repository
//! 5. Initialize state in Redis cache
//! 6. Return registered courier

use std::sync::Arc;

use thiserror::Error;

use crate::domain::ports::{
    CacheError, CachedCourierState, CommandHandlerWithResult, CourierCache, CourierRepository,
    RepositoryError,
};
use crate::domain::model::courier::{Courier, CourierError, CourierStatus};

use super::Command;

/// Errors that can occur during courier registration
#[derive(Debug, Error)]
pub enum RegisterCourierError {
    /// Email already registered
    #[error("Email already registered: {0}")]
    EmailExists(String),

    /// Phone already registered
    #[error("Phone already registered: {0}")]
    PhoneExists(String),

    /// Invalid courier data
    #[error("Invalid courier data: {0}")]
    InvalidData(String),

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    /// Cache error
    #[error("Cache error: {0}")]
    CacheError(#[from] CacheError),

    /// Domain error
    #[error("Domain error: {0}")]
    DomainError(#[from] CourierError),
}

/// Response from registering a courier
#[derive(Debug, Clone)]
pub struct Response {
    /// The registered courier
    pub courier: Courier,
}

/// Register Courier Handler
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
    type Error = RegisterCourierError;

    /// Handle the RegisterCourier command
    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        // 1. Check if email already exists
        if self.repository.email_exists(&cmd.email).await? {
            return Err(RegisterCourierError::EmailExists(cmd.email));
        }

        // 2. Check if phone already exists
        if self.repository.phone_exists(&cmd.phone).await? {
            return Err(RegisterCourierError::PhoneExists(cmd.phone));
        }

        // 3. Create Courier aggregate
        let mut builder = Courier::builder(
            cmd.name,
            cmd.phone,
            cmd.email,
            cmd.transport_type,
            cmd.max_distance_km,
            cmd.work_zone.clone(),
            cmd.work_hours,
        );

        if let Some(token) = cmd.push_token {
            builder = builder.with_push_token(token);
        }

        let courier = builder.build()?;

        // 4. Save to PostgreSQL repository
        self.repository.save(&courier).await?;

        // 5. Initialize state in Redis cache
        let cached_state = CachedCourierState {
            status: CourierStatus::Unavailable,
            current_load: 0,
            max_load: courier.max_load(),
            rating: 0.0,
            successful_deliveries: 0,
            failed_deliveries: 0,
        };

        self.cache
            .initialize_state(courier.id().0, cached_state, &cmd.work_zone)
            .await?;

        // 6. Return registered courier
        Ok(Response { courier })
    }
}

#[cfg(test)]
mod tests {
    // TODO: Add unit tests with mock repository and cache
}
