//! Courier Location Entity
//!
//! Represents a courier's current GPS location with metadata.
//! Used for real-time tracking of courier positions.

use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::domain::model::vo::location::Location;

/// Maximum allowed speed in km/h
pub const MAX_SPEED_KMH: f64 = 200.0;

/// Courier location entity for tracking
#[derive(Debug, Clone, PartialEq)]
pub struct CourierLocation {
    /// Courier ID
    courier_id: Uuid,
    /// GPS location (lat, lon, accuracy)
    location: Location,
    /// Timestamp when location was recorded
    timestamp: DateTime<Utc>,
    /// Speed in km/h (optional)
    speed: Option<f64>,
    /// Heading in degrees 0-360 (optional, 0 = North)
    heading: Option<f64>,
}

/// Errors for courier location operations
#[derive(Debug, Clone, PartialEq)]
pub enum CourierLocationError {
    /// Invalid speed value
    InvalidSpeed(f64),
    /// Invalid heading value
    InvalidHeading(f64),
    /// Timestamp is too far in the future
    TimestampInFuture,
    /// Timestamp is too old
    TimestampTooOld,
}

impl std::fmt::Display for CourierLocationError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            CourierLocationError::InvalidSpeed(s) => {
                write!(f, "Invalid speed: {}. Must be >= 0 and <= {}", s, MAX_SPEED_KMH)
            }
            CourierLocationError::InvalidHeading(h) => {
                write!(f, "Invalid heading: {}. Must be between 0 and 360", h)
            }
            CourierLocationError::TimestampInFuture => {
                write!(f, "Timestamp is too far in the future")
            }
            CourierLocationError::TimestampTooOld => {
                write!(f, "Timestamp is too old (> 5 minutes)")
            }
        }
    }
}

impl std::error::Error for CourierLocationError {}

impl CourierLocation {
    /// Create a new courier location with validation
    ///
    /// # Arguments
    ///
    /// * `courier_id` - The courier's UUID
    /// * `location` - GPS location (already validated)
    /// * `timestamp` - When the location was recorded
    /// * `speed` - Optional speed in km/h
    /// * `heading` - Optional heading in degrees (0-360)
    ///
    /// # Errors
    ///
    /// Returns error if speed, heading, or timestamp are invalid
    pub fn new(
        courier_id: Uuid,
        location: Location,
        timestamp: DateTime<Utc>,
        speed: Option<f64>,
        heading: Option<f64>,
    ) -> Result<Self, CourierLocationError> {
        // Validate speed
        if let Some(s) = speed {
            if s < 0.0 || s > MAX_SPEED_KMH {
                return Err(CourierLocationError::InvalidSpeed(s));
            }
        }

        // Validate heading
        if let Some(h) = heading {
            if !(0.0..=360.0).contains(&h) {
                return Err(CourierLocationError::InvalidHeading(h));
            }
        }

        // Validate timestamp (not too far in future, not too old)
        let now = Utc::now();
        let diff_secs = (timestamp - now).num_seconds();
        
        // Allow 60 seconds into the future (clock drift)
        if diff_secs > 60 {
            return Err(CourierLocationError::TimestampInFuture);
        }
        
        // Reject locations older than 5 minutes
        if diff_secs < -300 {
            return Err(CourierLocationError::TimestampTooOld);
        }

        Ok(Self {
            courier_id,
            location,
            timestamp,
            speed,
            heading,
        })
    }

    /// Create from stored data (no timestamp validation)
    ///
    /// Used when loading from database where timestamp validation is not needed
    pub fn from_stored(
        courier_id: Uuid,
        location: Location,
        timestamp: DateTime<Utc>,
        speed: Option<f64>,
        heading: Option<f64>,
    ) -> Result<Self, CourierLocationError> {
        // Validate speed
        if let Some(s) = speed {
            if s < 0.0 || s > MAX_SPEED_KMH {
                return Err(CourierLocationError::InvalidSpeed(s));
            }
        }

        // Validate heading
        if let Some(h) = heading {
            if !(0.0..=360.0).contains(&h) {
                return Err(CourierLocationError::InvalidHeading(h));
            }
        }

        Ok(Self {
            courier_id,
            location,
            timestamp,
            speed,
            heading,
        })
    }

    // === Getters ===

    /// Get the courier ID
    pub fn courier_id(&self) -> Uuid {
        self.courier_id
    }

    /// Get the location
    pub fn location(&self) -> &Location {
        &self.location
    }

    /// Get latitude
    pub fn latitude(&self) -> f64 {
        self.location.latitude()
    }

    /// Get longitude
    pub fn longitude(&self) -> f64 {
        self.location.longitude()
    }

    /// Get accuracy in meters
    pub fn accuracy(&self) -> f64 {
        self.location.accuracy()
    }

    /// Get the timestamp
    pub fn timestamp(&self) -> DateTime<Utc> {
        self.timestamp
    }

    /// Get speed in km/h
    pub fn speed(&self) -> Option<f64> {
        self.speed
    }

    /// Get heading in degrees
    pub fn heading(&self) -> Option<f64> {
        self.heading
    }

    /// Calculate distance to another courier location in kilometers
    pub fn distance_to(&self, other: &CourierLocation) -> f64 {
        self.location.distance_to(other.location())
    }

    /// Check if this location is within a certain radius of another
    pub fn is_within(&self, other: &CourierLocation, radius_km: f64) -> bool {
        self.distance_to(other) <= radius_km
    }

    /// Check if the location is stale (older than given seconds)
    pub fn is_stale(&self, max_age_secs: i64) -> bool {
        let age = Utc::now() - self.timestamp;
        age.num_seconds() > max_age_secs
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_location() -> Location {
        Location::new(52.52, 13.405, 10.0).unwrap()
    }

    #[test]
    fn test_courier_location_creation() {
        let courier_id = Uuid::new_v4();
        let location = create_location();
        let now = Utc::now();

        let courier_loc = CourierLocation::new(
            courier_id,
            location,
            now,
            Some(35.0),
            Some(180.0),
        );

        assert!(courier_loc.is_ok());
        let loc = courier_loc.unwrap();
        assert_eq!(loc.courier_id(), courier_id);
        assert_eq!(loc.latitude(), 52.52);
        assert_eq!(loc.longitude(), 13.405);
        assert_eq!(loc.speed(), Some(35.0));
        assert_eq!(loc.heading(), Some(180.0));
    }

    #[test]
    fn test_courier_location_without_optional_fields() {
        let courier_id = Uuid::new_v4();
        let location = create_location();
        let now = Utc::now();

        let courier_loc = CourierLocation::new(courier_id, location, now, None, None);

        assert!(courier_loc.is_ok());
        let loc = courier_loc.unwrap();
        assert_eq!(loc.speed(), None);
        assert_eq!(loc.heading(), None);
    }

    #[test]
    fn test_invalid_speed() {
        let courier_id = Uuid::new_v4();
        let location = create_location();
        let now = Utc::now();

        // Negative speed
        let result = CourierLocation::new(courier_id, location, now, Some(-1.0), None);
        assert!(matches!(result, Err(CourierLocationError::InvalidSpeed(_))));

        // Speed too high
        let result = CourierLocation::new(courier_id, location, now, Some(201.0), None);
        assert!(matches!(result, Err(CourierLocationError::InvalidSpeed(_))));
    }

    #[test]
    fn test_invalid_heading() {
        let courier_id = Uuid::new_v4();
        let location = create_location();
        let now = Utc::now();

        // Negative heading
        let result = CourierLocation::new(courier_id, location, now, None, Some(-1.0));
        assert!(matches!(result, Err(CourierLocationError::InvalidHeading(_))));

        // Heading > 360
        let result = CourierLocation::new(courier_id, location, now, None, Some(361.0));
        assert!(matches!(result, Err(CourierLocationError::InvalidHeading(_))));
    }

    #[test]
    fn test_boundary_heading_values() {
        let courier_id = Uuid::new_v4();
        let location = create_location();
        let now = Utc::now();

        // 0 degrees (North)
        assert!(CourierLocation::new(courier_id, location, now, None, Some(0.0)).is_ok());

        // 360 degrees (also North)
        assert!(CourierLocation::new(courier_id, location, now, None, Some(360.0)).is_ok());
    }

    #[test]
    fn test_distance_calculation() {
        let courier1 = CourierLocation::new(
            Uuid::new_v4(),
            Location::new(52.52, 13.405, 10.0).unwrap(),
            Utc::now(),
            None,
            None,
        ).unwrap();

        let courier2 = CourierLocation::new(
            Uuid::new_v4(),
            Location::new(48.1351, 11.582, 10.0).unwrap(), // Munich
            Utc::now(),
            None,
            None,
        ).unwrap();

        let distance = courier1.distance_to(&courier2);
        // Berlin to Munich is approximately 504 km
        assert!((distance - 504.0).abs() < 20.0);
    }

    #[test]
    fn test_is_within() {
        let now = Utc::now();
        let courier1 = CourierLocation::new(
            Uuid::new_v4(),
            Location::new(52.52, 13.405, 10.0).unwrap(),
            now,
            None,
            None,
        ).unwrap();

        let courier2 = CourierLocation::new(
            Uuid::new_v4(),
            Location::new(52.521, 13.406, 10.0).unwrap(),
            now,
            None,
            None,
        ).unwrap();

        assert!(courier1.is_within(&courier2, 1.0)); // Within 1km
        assert!(!courier1.is_within(&courier2, 0.001)); // Not within 1 meter
    }

    #[test]
    fn test_from_stored() {
        let courier_id = Uuid::new_v4();
        let location = create_location();
        // Old timestamp - would fail new() but should pass from_stored()
        let old_timestamp = Utc::now() - chrono::Duration::hours(1);

        let result = CourierLocation::from_stored(
            courier_id,
            location,
            old_timestamp,
            Some(50.0),
            Some(90.0),
        );

        assert!(result.is_ok());
    }
}
