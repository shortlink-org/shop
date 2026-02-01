//! Update Courier Location Command
//!
//! Data structure representing the command to update a courier's location.

use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::domain::model::vo::location::Location;

/// Command to update a courier's location
#[derive(Debug, Clone)]
pub struct Command {
    /// The courier's ID
    pub courier_id: Uuid,
    /// The new location (coordinates + accuracy)
    pub location: Location,
    /// Timestamp when location was recorded
    pub timestamp: DateTime<Utc>,
    /// Speed in km/h (optional)
    pub speed: Option<f64>,
    /// Heading in degrees 0-360 (optional)
    pub heading: Option<f64>,
}

impl Command {
    /// Create a new UpdateCourierLocation command
    pub fn new(
        courier_id: Uuid,
        location: Location,
        timestamp: DateTime<Utc>,
        speed: Option<f64>,
        heading: Option<f64>,
    ) -> Self {
        Self {
            courier_id,
            location,
            timestamp,
            speed,
            heading,
        }
    }

    /// Create command from millisecond timestamp
    pub fn from_timestamp_ms(
        courier_id: Uuid,
        location: Location,
        timestamp_ms: i64,
        speed: Option<f64>,
        heading: Option<f64>,
    ) -> Self {
        let timestamp = DateTime::from_timestamp_millis(timestamp_ms)
            .unwrap_or_else(Utc::now);
        Self::new(courier_id, location, timestamp, speed, heading)
    }
}
