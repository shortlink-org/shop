//! Courier Location Module
//!
//! Contains entities for courier location tracking:
//! - `CourierLocation` - current courier position
//! - `LocationHistoryEntry` - historical location record

mod entity;
mod history;

pub use entity::{CourierLocation, CourierLocationError, MAX_SPEED_KMH};
pub use history::{LocationHistoryEntry, TimeRange};
