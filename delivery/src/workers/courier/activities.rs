//! Courier Activities
//!
//! Temporal activities for courier operations.
//! Activities are thin wrappers that delegate to use cases.
//!
//! These activities are registered with the Temporal worker in `runner.rs`
//! and called from courier workflows.

use std::sync::Arc;

use thiserror::Error;
use tracing::warn;
use uuid::Uuid;

use crate::domain::model::courier::{Courier, CourierStatus, WorkHours};
use crate::domain::model::vo::TransportType;
use crate::domain::ports::{
    CommandHandlerWithResult, CourierCache, CourierRepository, QueryHandler,
};
use crate::usecases::courier::command::{
    AcceptPackageCommand, AcceptPackageError, AcceptPackageHandler, CompleteCourierDeliveryCommand,
    CompleteCourierDeliveryError, CompleteCourierDeliveryHandler, RegisterCommand, RegisterHandler,
};
use crate::usecases::courier::query::get_pool::{Handler as GetPoolHandler, Query as GetPoolQuery};

/// Errors from courier activities
#[derive(Debug, Error)]
pub enum CourierActivityError {
    #[error("Use case error: {0}")]
    UseCaseError(String),

    #[error("Courier not found: {0}")]
    NotFound(Uuid),

    #[error("Invalid operation: {0}")]
    InvalidOperation(String),
}

/// Courier Activities - thin wrappers around use cases
///
/// Following hexagonal architecture, activities should:
/// - Call use cases for business operations
/// - NOT contain business logic
/// - NOT directly access repositories (delegate to use cases)
pub struct CourierActivities<R, C>
where
    R: CourierRepository + 'static,
    C: CourierCache + 'static,
{
    register_handler: Arc<RegisterHandler<R, C>>,
    get_pool_handler: Arc<GetPoolHandler<R, C>>,
    accept_package_handler: Arc<AcceptPackageHandler<R, C>>,
    complete_delivery_handler: Arc<CompleteCourierDeliveryHandler<R, C>>,
    repository: Arc<R>,
    cache: Arc<C>,
}

impl<R, C> CourierActivities<R, C>
where
    R: CourierRepository + Send + Sync + 'static,
    C: CourierCache + Send + Sync + 'static,
{
    /// Create new courier activities
    pub fn new(
        register_handler: Arc<RegisterHandler<R, C>>,
        get_pool_handler: Arc<GetPoolHandler<R, C>>,
        accept_package_handler: Arc<AcceptPackageHandler<R, C>>,
        complete_delivery_handler: Arc<CompleteCourierDeliveryHandler<R, C>>,
        repository: Arc<R>,
        cache: Arc<C>,
    ) -> Self {
        Self {
            register_handler,
            get_pool_handler,
            accept_package_handler,
            complete_delivery_handler,
            repository,
            cache,
        }
    }

    async fn save_courier_and_refresh_cache(
        &self,
        courier: &Courier,
        operation: &'static str,
    ) -> Result<(), CourierActivityError> {
        self.repository
            .save(courier)
            .await
            .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))?;

        if let Err(err) = self.cache.cache(courier).await {
            warn!(
                courier_id = %courier.id().0,
                operation,
                error = %err,
                "Courier cache refresh failed after repository save; treating as non-fatal for Temporal activity"
            );
        }

        Ok(())
    }

    // =========================================================================
    // Activity: Register Courier
    // =========================================================================

    /// Register a new courier in the system
    ///
    /// This activity delegates to RegisterHandler.
    #[allow(clippy::too_many_arguments)]
    pub async fn register_courier(
        &self,
        name: String,
        phone: String,
        email: String,
        transport_type: TransportType,
        max_distance_km: f64,
        work_zone: String,
        work_hours: WorkHours,
        push_token: Option<String>,
    ) -> Result<Courier, CourierActivityError> {
        let command = RegisterCommand::new(
            name,
            phone,
            email,
            transport_type,
            max_distance_km,
            work_zone,
            work_hours,
            push_token,
        );

        self.register_handler
            .handle(command)
            .await
            .map(|r| r.courier)
            .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))
    }

    // =========================================================================
    // Activity: Get Free Couriers
    // =========================================================================

    /// Get free couriers in a zone
    ///
    /// This activity delegates to GetPoolHandler.
    pub async fn get_free_couriers_in_zone(
        &self,
        zone: &str,
    ) -> Result<Vec<Courier>, CourierActivityError> {
        let query = GetPoolQuery::free_in_zone(zone);

        self.get_pool_handler
            .handle(query)
            .await
            .map(|r| r.couriers.into_iter().map(|c| c.courier).collect())
            .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))
    }

    // =========================================================================
    // Activity: Update Courier Status
    // =========================================================================

    /// Update courier status (go online/offline)
    ///
    /// Updates both repository and cache.
    pub async fn update_status(
        &self,
        courier_id: Uuid,
        status: CourierStatus,
    ) -> Result<(), CourierActivityError> {
        let mut courier = self
            .repository
            .find_by_id(courier_id)
            .await
            .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))?
            .ok_or(CourierActivityError::NotFound(courier_id))?;

        match status {
            CourierStatus::Free => courier
                .go_online()
                .map_err(|e| CourierActivityError::InvalidOperation(e.to_string()))?,
            CourierStatus::Unavailable => courier
                .go_offline()
                .map_err(|e| CourierActivityError::InvalidOperation(e.to_string()))?,
            CourierStatus::Archived => courier
                .archive()
                .map_err(|e| CourierActivityError::InvalidOperation(e.to_string()))?,
            CourierStatus::Busy => {
                return Err(CourierActivityError::InvalidOperation(
                    "Busy status is derived from active assignments".to_string(),
                ))
            }
        }

        self.save_courier_and_refresh_cache(&courier, "update_status")
            .await
    }

    // =========================================================================
    // Activity: Accept Package
    // =========================================================================

    /// Accept a package assignment
    ///
    /// Updates courier load in cache.
    pub async fn accept_package(&self, courier_id: Uuid) -> Result<(), CourierActivityError> {
        self.accept_package_handler
            .handle(AcceptPackageCommand::new(courier_id))
            .await
            .map(|_| ())
            .map_err(Self::map_accept_package_error)
    }

    // =========================================================================
    // Activity: Complete Delivery
    // =========================================================================

    /// Complete a delivery
    ///
    /// Updates courier load and stats in cache.
    pub async fn complete_delivery(
        &self,
        courier_id: Uuid,
        success: bool,
    ) -> Result<(), CourierActivityError> {
        self.complete_delivery_handler
            .handle(CompleteCourierDeliveryCommand::new(courier_id, success))
            .await
            .map(|_| ())
            .map_err(Self::map_complete_delivery_error)
    }

    fn map_accept_package_error(error: AcceptPackageError) -> CourierActivityError {
        match error {
            AcceptPackageError::NotFound(courier_id) => CourierActivityError::NotFound(courier_id),
            AcceptPackageError::DomainError(error) => {
                CourierActivityError::InvalidOperation(error.to_string())
            }
            AcceptPackageError::RepositoryError(error) => {
                CourierActivityError::UseCaseError(error.to_string())
            }
        }
    }

    fn map_complete_delivery_error(error: CompleteCourierDeliveryError) -> CourierActivityError {
        match error {
            CompleteCourierDeliveryError::NotFound(courier_id) => {
                CourierActivityError::NotFound(courier_id)
            }
            CompleteCourierDeliveryError::DomainError(error) => {
                CourierActivityError::InvalidOperation(error.to_string())
            }
            CompleteCourierDeliveryError::RepositoryError(error) => {
                CourierActivityError::UseCaseError(error.to_string())
            }
        }
    }
}

// Note: Activity registration is done in `runner.rs` using the Temporal SDK.
// The methods here use domain/application types; `runner.rs` provides the
// Temporal-facing wrappers that serialize activity inputs/outputs as strings
// for compatibility with the current pre-alpha SDK.
