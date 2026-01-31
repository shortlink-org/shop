//! Package Aggregate
//!
//! Package represents a delivery package in the system.
//! It manages its own state transitions and business rules.

pub mod entity;
pub mod state;

pub use entity::*;
pub use state::*;
