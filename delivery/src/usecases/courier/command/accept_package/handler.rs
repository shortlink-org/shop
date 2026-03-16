//! Accept Package Handler
//!
//! Handles increasing courier load after assignment.

use std::sync::Arc;

use thiserror::Error;
use tracing::{info, warn};
use uuid::Uuid;

use crate::domain::model::courier::{CourierError, CourierStatus};
use crate::domain::ports::{
    CommandHandlerWithResult, CourierCache, CourierRepository, RepositoryError,
};

use super::Command;

/// Errors that can occur during package acceptance.
#[derive(Debug, Error)]
pub enum AcceptPackageError {
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

/// Response from accepting a package.
#[derive(Debug, Clone)]
pub struct Response {
    /// Courier ID.
    pub courier_id: Uuid,
    /// Current courier status after the update.
    pub status: CourierStatus,
    /// Current load after the update.
    pub current_load: u32,
}

/// Accept Package Handler.
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
    type Error = AcceptPackageError;

    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        let mut courier = self
            .courier_repo
            .find_by_id(cmd.courier_id)
            .await?
            .ok_or(AcceptPackageError::NotFound(cmd.courier_id))?;

        info!(
            courier_id = %cmd.courier_id,
            phase = "loaded",
            current_load = courier.current_load(),
            status = %courier.status(),
            "Courier loaded for accept_package"
        );

        courier.accept_package()?;
        self.courier_repo.save(&courier).await?;

        info!(
            courier_id = %cmd.courier_id,
            phase = "saved",
            current_load = courier.current_load(),
            status = %courier.status(),
            "Courier saved after accept_package"
        );

        if let Err(err) = self.courier_cache.cache(&courier).await {
            warn!(
                courier_id = %cmd.courier_id,
                phase = "cache_updated",
                error = %err,
                "Courier cache refresh failed after accept_package; treating as non-fatal"
            );
        } else {
            info!(
                courier_id = %cmd.courier_id,
                phase = "cache_updated",
                "Courier cache refreshed after accept_package"
            );
        }

        Ok(Response {
            courier_id: cmd.courier_id,
            status: courier.status(),
            current_load: courier.current_load(),
        })
    }
}
