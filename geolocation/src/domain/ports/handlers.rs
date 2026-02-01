//! Command and Query Handler traits
//!
//! Generic handler interfaces for CQRS pattern.

use async_trait::async_trait;

/// Command handler with result
#[async_trait]
pub trait CommandHandlerWithResult<C, R>: Send + Sync {
    type Error;

    async fn handle(&self, command: C) -> Result<R, Self::Error>;
}

/// Query handler
#[async_trait]
pub trait QueryHandler<Q, R>: Send + Sync {
    type Error;

    async fn handle(&self, query: Q) -> Result<R, Self::Error>;
}
