//! Save Location Command
//!
//! Command to save a courier's current location.

use chrono::{DateTime, Utc};
use uuid::Uuid;

/// Command to save courier's location
#[derive(Debug, Clone)]
pub struct Command {
    /// Courier ID
    pub courier_id: Uuid,
    /// Latitude
    pub latitude: f64,
    /// Longitude
    pub longitude: f64,
    /// GPS accuracy in meters
    pub accuracy: f64,
    /// Location timestamp
    pub timestamp: DateTime<Utc>,
    /// Speed in km/h (optional)
    pub speed: Option<f64>,
    /// Heading in degrees 0-360 (optional)
    pub heading: Option<f64>,
}

impl Command {
    /// Create a new SaveLocation command
    pub fn new(
        courier_id: Uuid,
        latitude: f64,
        longitude: f64,
        accuracy: f64,
        timestamp: DateTime<Utc>,
        speed: Option<f64>,
        heading: Option<f64>,
    ) -> Self {
        Self {
            courier_id,
            latitude,
            longitude,
            accuracy,
            timestamp,
            speed,
            heading,
        }
    }
}
