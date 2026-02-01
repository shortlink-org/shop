//! Value Objects
//!
//! Immutable objects defined by their attributes rather than identity.

pub mod coordinates;
pub mod location;

pub use coordinates::{Coordinates, CoordinatesError};
pub use location::{Location, LocationError};
