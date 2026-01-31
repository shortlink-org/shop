//! Register Courier Use Case
//!
//! Registers a new courier in the system.
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

use crate::boundary::ports::{CacheError, CachedCourierState, CourierCache, CourierRepository, RepositoryError};
use crate::domain::model::courier::{Courier, CourierError, CourierStatus, WorkHours};
use crate::domain::model::vo::TransportType;

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

/// Request to register a new courier
#[derive(Debug, Clone)]
pub struct RegisterCourierRequest {
    pub name: String,
    pub phone: String,
    pub email: String,
    pub transport_type: TransportType,
    pub max_distance_km: f64,
    pub work_zone: String,
    pub work_hours: WorkHours,
    pub push_token: Option<String>,
}

/// Response from registering a courier
#[derive(Debug, Clone)]
pub struct RegisterCourierResponse {
    pub courier: Courier,
}

/// Register Courier Use Case
pub struct RegisterCourierUseCase<R, C>
where
    R: CourierRepository,
    C: CourierCache,
{
    repository: Arc<R>,
    cache: Arc<C>,
}

impl<R, C> RegisterCourierUseCase<R, C>
where
    R: CourierRepository,
    C: CourierCache,
{
    /// Create a new use case instance
    pub fn new(repository: Arc<R>, cache: Arc<C>) -> Self {
        Self { repository, cache }
    }

    /// Execute the use case
    pub async fn execute(
        &self,
        request: RegisterCourierRequest,
    ) -> Result<RegisterCourierResponse, RegisterCourierError> {
        // 1. Check if email already exists
        if self.repository.email_exists(&request.email).await? {
            return Err(RegisterCourierError::EmailExists(request.email));
        }

        // 2. Check if phone already exists
        if self.repository.phone_exists(&request.phone).await? {
            return Err(RegisterCourierError::PhoneExists(request.phone));
        }

        // 3. Create Courier aggregate
        let mut builder = Courier::builder(
            request.name,
            request.phone,
            request.email,
            request.transport_type,
            request.max_distance_km,
            request.work_zone.clone(),
            request.work_hours,
        );

        if let Some(token) = request.push_token {
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
            .initialize_state(courier.id().0, cached_state, &request.work_zone)
            .await?;

        // 6. Return registered courier
        Ok(RegisterCourierResponse { courier })
    }
}

#[cfg(test)]
mod tests {
    // TODO: Add unit tests with mock repository and cache
}
