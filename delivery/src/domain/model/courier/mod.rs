//! Courier Aggregate
//!
//! Courier represents a delivery courier in the system.
//! It manages its own state, capacity, and availability rules.

pub mod entity;
pub mod state;

pub use entity::*;
pub use state::*;
