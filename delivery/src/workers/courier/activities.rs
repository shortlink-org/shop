//! Courier Activities
//!
//! Temporal activities for courier operations.
//! Activities are thin wrappers that delegate to use cases.
//!
//! These activities are registered with the Temporal worker in `runner.rs`
//! and called from courier workflows.

use std::sync::Arc;

use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::courier::{Courier, CourierStatus, WorkHours};
use crate::domain::model::vo::TransportType;
use crate::domain::ports::{CommandHandlerWithResult, CourierCache, CourierRepository, QueryHandler};
use crate::usecases::courier::command::register::{Command as RegisterCommand, Handler as RegisterHandler};
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
        repository: Arc<R>,
        cache: Arc<C>,
    ) -> Self {
        Self {
            register_handler,
            get_pool_handler,
            repository,
            cache,
        }
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
        // Get courier from repository
        let courier = self
            .repository
            .find_by_id(courier_id)
            .await
            .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))?
            .ok_or(CourierActivityError::NotFound(courier_id))?;

        // Update cache
        self.cache
            .set_status(courier_id, status, courier.work_zone())
            .await
            .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))?;

        Ok(())
    }

    // =========================================================================
    // Activity: Accept Package
    // =========================================================================

    /// Accept a package assignment
    ///
    /// Updates courier load in cache.
    pub async fn accept_package(&self, courier_id: Uuid) -> Result<(), CourierActivityError> {
        // Get current state from cache
        let state = self
            .cache
            .get_state(courier_id)
            .await
            .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))?
            .ok_or(CourierActivityError::NotFound(courier_id))?;

        // Check capacity
        if state.current_load >= state.max_load {
            return Err(CourierActivityError::InvalidOperation(
                "Courier at full capacity".to_string(),
            ));
        }

        // Update load
        let new_load = state.current_load + 1;
        self.cache
            .update_load(courier_id, new_load, state.max_load)
            .await
            .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))?;

        // If at capacity, update status to Busy
        if new_load >= state.max_load {
            let courier = self
                .repository
                .find_by_id(courier_id)
                .await
                .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))?
                .ok_or(CourierActivityError::NotFound(courier_id))?;

            self.cache
                .set_status(courier_id, CourierStatus::Busy, courier.work_zone())
                .await
                .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))?;
        }

        Ok(())
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
        // Get current state from cache
        let state = self
            .cache
            .get_state(courier_id)
            .await
            .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))?
            .ok_or(CourierActivityError::NotFound(courier_id))?;

        // Update load
        let new_load = state.current_load.saturating_sub(1);
        self.cache
            .update_load(courier_id, new_load, state.max_load)
            .await
            .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))?;

        // Update stats
        let (successful, failed) = if success {
            (state.successful_deliveries + 1, state.failed_deliveries)
        } else {
            (state.successful_deliveries, state.failed_deliveries + 1)
        };

        let total = successful + failed;
        let rating = if total > 0 {
            (successful as f64 / total as f64) * 5.0
        } else {
            0.0
        };

        self.cache
            .update_stats(courier_id, rating, successful, failed)
            .await
            .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))?;

        // If was at capacity and now has space, update status to Free
        if state.current_load >= state.max_load && new_load < state.max_load {
            let courier = self
                .repository
                .find_by_id(courier_id)
                .await
                .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))?
                .ok_or(CourierActivityError::NotFound(courier_id))?;

            self.cache
                .set_status(courier_id, CourierStatus::Free, courier.work_zone())
                .await
                .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))?;
        }

        Ok(())
    }
}

// Note: Activity registration is done in `runner.rs` using the Temporal SDK.
// Each activity method above is wrapped and registered with the worker there.
//
// Activity input/output types use simple strings for serialization compatibility
// with the pre-alpha Temporal SDK.
