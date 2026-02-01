//! Domain Models
//!
//! This module contains all domain models including:
//! - Aggregates (Package, Courier)
//! - Entities (CourierLocation)
//! - Value Objects (Location, Address, etc.)
//! - Proto-generated models (delivery/)

pub mod courier;
pub mod courier_location;
pub mod package;
pub mod vo;

// Re-exports for convenience
pub use courier_location::{CourierLocation, CourierLocationError, LocationHistoryEntry, TimeRange};

// Proto generated code will be here after build
// pub mod delivery;
