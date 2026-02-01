//! Domain Models
//!
//! Aggregates, Entities, and Value Objects for the Geolocation service.

pub mod courier_location;
pub mod geofence;
pub mod location_history;
pub mod vo;

pub use courier_location::CourierLocation;
pub use geofence::{Geofence, GeofenceError, GeofenceShape, GeofenceTrigger};
pub use location_history::{LocationHistoryEntry, TimeRange};
pub use vo::{Coordinates, CoordinatesError, Location, LocationError};
