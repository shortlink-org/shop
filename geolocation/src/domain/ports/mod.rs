//! Domain Ports
//!
//! Interfaces (traits) for external dependencies.

mod handlers;
mod location_cache;
mod location_repository;

pub use handlers::{CommandHandlerWithResult, QueryHandler};
pub use location_cache::{CacheError, LocationCache};
pub use location_repository::{GeofenceRepository, LocationRepository, RepositoryError};
