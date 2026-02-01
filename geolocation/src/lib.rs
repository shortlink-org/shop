//! Geolocation Service
//!
//! Provides location tracking and geofencing capabilities for couriers.
//!
//! ## Architecture
//!
//! The service follows Clean Architecture principles:
//!
//! - **domain/model** - Entities, Aggregates, and Value Objects
//! - **domain/ports** - Repository and service interfaces
//! - **usecases** - Application layer with CQRS commands and queries
//!
//! ## Use Cases
//!
//! ### Commands (write operations)
//! - `SaveLocation` - Save courier's current location
//! - `CreateGeofence` - Create a new geofence
//!
//! ### Queries (read operations)
//! - `GetCourierLocations` - Get current locations for one or more couriers
//! - `GetLocationHistory` - Get location history for a courier

pub mod domain;
pub mod usecases;

// Re-export commonly used types
pub use domain::model::{
    Coordinates, CoordinatesError, CourierLocation, Geofence, GeofenceError, GeofenceShape,
    GeofenceTrigger, Location, LocationError, LocationHistoryEntry, TimeRange,
};

pub use domain::ports::{
    CacheError, CommandHandlerWithResult, GeofenceRepository, LocationCache, LocationRepository,
    QueryHandler, RepositoryError,
};
