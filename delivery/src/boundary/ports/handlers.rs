//! Handler Traits
//!
//! Defines the contracts for command and query handlers following CQRS pattern.
//! These traits provide a consistent interface for all use case handlers.

use std::future::Future;

/// CommandHandler handles commands that modify state without returning a result.
/// C = Command type
pub trait CommandHandler<C>: Send + Sync {
    /// The error type returned by this handler
    type Error;

    /// Handle the command
    fn handle(&self, cmd: C) -> impl Future<Output = Result<(), Self::Error>> + Send;
}

/// CommandHandlerWithResult handles commands that modify state and return a result.
/// C = Command type, R = Result type
pub trait CommandHandlerWithResult<C, R>: Send + Sync {
    /// The error type returned by this handler
    type Error;

    /// Handle the command and return a result
    fn handle(&self, cmd: C) -> impl Future<Output = Result<R, Self::Error>> + Send;
}

/// QueryHandler handles read-only queries.
/// Q = Query type, R = Result type
pub trait QueryHandler<Q, R>: Send + Sync {
    /// The error type returned by this handler
    type Error;

    /// Handle the query and return a result
    fn handle(&self, query: Q) -> impl Future<Output = Result<R, Self::Error>> + Send;
}

/// EventHandler handles domain events (reactions to facts).
/// E = Event type
pub trait EventHandler<E>: Send + Sync {
    /// The error type returned by this handler
    type Error;

    /// Handle the event
    fn handle(&self, event: E) -> impl Future<Output = Result<(), Self::Error>> + Send;
}
