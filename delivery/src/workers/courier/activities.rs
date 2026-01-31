//! Courier Activities
//!
//! Temporal activities for courier operations.
//! Activities are thin wrappers that delegate to use cases.
//!
//! NOTE: This is a placeholder implementation. The actual Temporal Rust SDK
//! is still in development. This code demonstrates the intended patterns.

use std::sync::Arc;

use thiserror::Error;
use uuid::Uuid;

use crate::boundary::ports::{CourierCache, CourierRepository};
use crate::domain::model::courier::{Courier, CourierStatus, WorkHours};
use crate::domain::model::vo::TransportType;
use crate::usecases::{
    CourierFilter, GetCourierPoolUseCase, RegisterCourierRequest, RegisterCourierUseCase,
};

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
    register_uc: Arc<RegisterCourierUseCase<R, C>>,
    get_pool_uc: Arc<GetCourierPoolUseCase<R, C>>,
    repository: Arc<R>,
    cache: Arc<C>,
}

impl<R, C> CourierActivities<R, C>
where
    R: CourierRepository + 'static,
    C: CourierCache + 'static,
{
    /// Create new courier activities
    pub fn new(
        register_uc: Arc<RegisterCourierUseCase<R, C>>,
        get_pool_uc: Arc<GetCourierPoolUseCase<R, C>>,
        repository: Arc<R>,
        cache: Arc<C>,
    ) -> Self {
        Self {
            register_uc,
            get_pool_uc,
            repository,
            cache,
        }
    }

    // =========================================================================
    // Activity: Register Courier
    // =========================================================================

    /// Register a new courier in the system
    ///
    /// This activity delegates to RegisterCourierUseCase.
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
        let request = RegisterCourierRequest {
            name,
            phone,
            email,
            transport_type,
            max_distance_km,
            work_zone,
            work_hours,
            push_token,
        };

        self.register_uc
            .execute(request)
            .await
            .map(|r| r.courier)
            .map_err(|e| CourierActivityError::UseCaseError(e.to_string()))
    }

    // =========================================================================
    // Activity: Get Free Couriers
    // =========================================================================

    /// Get free couriers in a zone
    ///
    /// This activity delegates to GetCourierPoolUseCase.
    pub async fn get_free_couriers_in_zone(
        &self,
        zone: &str,
    ) -> Result<Vec<Courier>, CourierActivityError> {
        let filter = CourierFilter::free_in_zone(zone);

        self.get_pool_uc
            .execute(filter)
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

// =============================================================================
// Temporal Activity Registration (Placeholder)
//
// When Temporal Rust SDK is stable, register activities like this:
//
// ```rust
// use temporal_sdk::Worker;
//
// pub fn register_courier_activities<R, C>(
//     worker: &mut Worker,
//     activities: Arc<CourierActivities<R, C>>,
// ) where
//     R: CourierRepository + 'static,
//     C: CourierCache + 'static,
// {
//     worker.register_activity("register_courier", {
//         let acts = activities.clone();
//         move |input: RegisterCourierInput| {
//             let acts = acts.clone();
//             async move {
//                 acts.register_courier(
//                     input.name,
//                     input.phone,
//                     input.email,
//                     input.transport_type,
//                     input.max_distance_km,
//                     input.work_zone,
//                     input.work_hours,
//                     input.push_token,
//                 ).await
//             }
//         }
//     });
//
//     worker.register_activity("get_free_couriers", {
//         let acts = activities.clone();
//         move |zone: String| {
//             let acts = acts.clone();
//             async move { acts.get_free_couriers_in_zone(&zone).await }
//         }
//     });
//
//     // ... register other activities
// }
// ```
// =============================================================================
