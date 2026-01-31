//! Delivery Activities
//!
//! Temporal activities for delivery operations.
//!
//! NOTE: This is a placeholder implementation. The actual Temporal Rust SDK
//! is still in development.

use std::sync::Arc;

use thiserror::Error;
use uuid::Uuid;

use crate::boundary::ports::{CourierCache, CourierRepository};
use crate::domain::services::dispatch::{CourierForDispatch, DispatchResult, DispatchService, PackageForDispatch};
use crate::usecases::{CourierFilter, GetCourierPoolUseCase};

/// Errors from delivery activities
#[derive(Debug, Error)]
pub enum DeliveryActivityError {
    #[error("Use case error: {0}")]
    UseCaseError(String),

    #[error("No couriers available in zone: {0}")]
    NoCouriersAvailable(String),

    #[error("Courier not found: {0}")]
    CourierNotFound(Uuid),

    #[error("Dispatch failed: {0}")]
    DispatchFailed(String),
}

/// Delivery Activities - thin wrappers around use cases
pub struct DeliveryActivities<R, C>
where
    R: CourierRepository + 'static,
    C: CourierCache + 'static,
{
    get_pool_uc: Arc<GetCourierPoolUseCase<R, C>>,
    cache: Arc<C>,
}

impl<R, C> DeliveryActivities<R, C>
where
    R: CourierRepository + 'static,
    C: CourierCache + 'static,
{
    /// Create new delivery activities
    pub fn new(get_pool_uc: Arc<GetCourierPoolUseCase<R, C>>, cache: Arc<C>) -> Self {
        Self { get_pool_uc, cache }
    }

    // =========================================================================
    // Activity: Get Free Couriers for Dispatch
    // =========================================================================

    /// Get free couriers in a zone, ready for dispatch
    pub async fn get_free_couriers_for_dispatch(
        &self,
        zone: &str,
    ) -> Result<Vec<CourierForDispatch>, DeliveryActivityError> {
        let filter = CourierFilter::free_in_zone(zone);

        let result = self
            .get_pool_uc
            .execute(filter)
            .await
            .map_err(|e| DeliveryActivityError::UseCaseError(e.to_string()))?;

        // Convert to CourierForDispatch
        let couriers: Vec<CourierForDispatch> = result
            .couriers
            .into_iter()
            .filter_map(|cws| {
                let state = cws.state?;
                Some(CourierForDispatch {
                    id: cws.courier.id().to_string(),
                    status: state.status,
                    transport_type: cws.courier.transport_type(),
                    max_distance_km: cws.courier.max_distance_km(),
                    capacity: crate::domain::model::courier::CourierCapacity::new(state.max_load),
                    current_location: None, // Will be fetched separately from Geolocation Service
                    rating: state.rating,
                    work_zone: cws.courier.work_zone().to_string(),
                })
            })
            .collect();

        Ok(couriers)
    }

    // =========================================================================
    // Activity: Find Nearest Courier
    // =========================================================================

    /// Find the nearest courier for a package (uses domain service)
    ///
    /// Note: This activity receives couriers with locations already fetched
    /// from the Geolocation Service. The actual location fetching should be
    /// done in a separate activity.
    pub fn find_nearest_courier(
        &self,
        couriers: &[CourierForDispatch],
        package: &PackageForDispatch,
    ) -> Option<DispatchResult> {
        DispatchService::find_nearest_courier(couriers, package)
    }

    // =========================================================================
    // Activity: Assign Order to Courier
    // =========================================================================

    /// Assign an order to a courier
    ///
    /// Updates courier state in cache.
    pub async fn assign_order(
        &self,
        courier_id: Uuid,
        _order_id: Uuid,
    ) -> Result<(), DeliveryActivityError> {
        // Get current state
        let state = self
            .cache
            .get_state(courier_id)
            .await
            .map_err(|e| DeliveryActivityError::UseCaseError(e.to_string()))?
            .ok_or(DeliveryActivityError::CourierNotFound(courier_id))?;

        // Increment load
        let new_load = state.current_load + 1;
        self.cache
            .update_load(courier_id, new_load, state.max_load)
            .await
            .map_err(|e| DeliveryActivityError::UseCaseError(e.to_string()))?;

        // If at capacity, mark as busy
        // Note: We don't have the work_zone here, so we'd need to look it up
        // In practice, this should be passed as a parameter or looked up from repository

        Ok(())
    }

    // =========================================================================
    // Activity: Complete Delivery
    // =========================================================================

    /// Complete a delivery (success or failure)
    pub async fn complete_delivery(
        &self,
        courier_id: Uuid,
        _order_id: Uuid,
        success: bool,
    ) -> Result<(), DeliveryActivityError> {
        // Get current state
        let state = self
            .cache
            .get_state(courier_id)
            .await
            .map_err(|e| DeliveryActivityError::UseCaseError(e.to_string()))?
            .ok_or(DeliveryActivityError::CourierNotFound(courier_id))?;

        // Decrement load
        let new_load = state.current_load.saturating_sub(1);
        self.cache
            .update_load(courier_id, new_load, state.max_load)
            .await
            .map_err(|e| DeliveryActivityError::UseCaseError(e.to_string()))?;

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
            .map_err(|e| DeliveryActivityError::UseCaseError(e.to_string()))?;

        Ok(())
    }
}
