//! Update Courier Location Command
//!
//! Data structure representing the command to update a courier's location.

use uuid::Uuid;

use crate::domain::model::vo::location::Location;

/// Command to update a courier's location
#[derive(Debug, Clone)]
pub struct Command {
    /// The courier's ID
    pub courier_id: Uuid,
    /// The new location
    pub location: Location,
    /// Timestamp of the location update (Unix timestamp in milliseconds)
    pub timestamp_ms: i64,
}

impl Command {
    /// Create a new UpdateCourierLocation command
    pub fn new(courier_id: Uuid, location: Location, timestamp_ms: i64) -> Self {
        Self {
            courier_id,
            location,
            timestamp_ms,
        }
    }
}
