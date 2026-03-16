//! Complete Courier Delivery Handler
//!
//! Handles decreasing courier load and updating delivery stats.

use std::sync::Arc;

use thiserror::Error;
use tracing::{info, warn};
use uuid::Uuid;

use crate::domain::model::courier::{CourierError, CourierStatus};
use crate::domain::ports::{
    CommandHandlerWithResult, CourierCache, CourierRepository, RepositoryError,
};

use super::Command;

/// Errors that can occur during courier delivery completion.
#[derive(Debug, Error)]
pub enum CompleteCourierDeliveryError {
    /// Courier not found.
    #[error("Courier not found: {0}")]
    NotFound(Uuid),

    /// Repository error.
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    /// Domain error.
    #[error("Domain error: {0}")]
    DomainError(#[from] CourierError),
}

/// Response from completing a courier delivery.
#[derive(Debug, Clone)]
pub struct Response {
    /// Courier ID.
    pub courier_id: Uuid,
    /// Current courier status after the update.
    pub status: CourierStatus,
    /// Current load after the update.
    pub current_load: u32,
}

/// Complete Courier Delivery Handler.
pub struct Handler<R, C>
where
    R: CourierRepository,
    C: CourierCache,
{
    courier_repo: Arc<R>,
    courier_cache: Arc<C>,
}

impl<R, C> Handler<R, C>
where
    R: CourierRepository,
    C: CourierCache,
{
    /// Create a new handler instance.
    pub fn new(courier_repo: Arc<R>, courier_cache: Arc<C>) -> Self {
        Self {
            courier_repo,
            courier_cache,
        }
    }
}

impl<R, C> CommandHandlerWithResult<Command, Response> for Handler<R, C>
where
    R: CourierRepository + Send + Sync,
    C: CourierCache + Send + Sync,
{
    type Error = CompleteCourierDeliveryError;

    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        let mut courier = self
            .courier_repo
            .find_by_id(cmd.courier_id)
            .await?
            .ok_or(CompleteCourierDeliveryError::NotFound(cmd.courier_id))?;

        info!(
            courier_id = %cmd.courier_id,
            success = cmd.success,
            phase = "loaded",
            current_load = courier.current_load(),
            status = %courier.status(),
            "Courier loaded for complete_delivery"
        );

        courier.complete_delivery(cmd.success)?;
        self.courier_repo.save(&courier).await?;

        info!(
            courier_id = %cmd.courier_id,
            success = cmd.success,
            phase = "saved",
            current_load = courier.current_load(),
            status = %courier.status(),
            "Courier saved after complete_delivery"
        );

        if let Err(err) = self.courier_cache.cache(&courier).await {
            warn!(
                courier_id = %cmd.courier_id,
                success = cmd.success,
                phase = "cache_updated",
                error = %err,
                "Courier cache refresh failed after complete_delivery; treating as non-fatal"
            );
        } else {
            info!(
                courier_id = %cmd.courier_id,
                success = cmd.success,
                phase = "cache_updated",
                "Courier cache refreshed after complete_delivery"
            );
        }

        Ok(Response {
            courier_id: cmd.courier_id,
            status: courier.status(),
            current_load: courier.current_load(),
        })
    }
}
