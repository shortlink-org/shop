//! Location Value Object
//!
//! Represents a GPS location with coordinates, accuracy, and optional metadata.

use chrono::{DateTime, Utc};
use thiserror::Error;

use super::coordinates::{Coordinates, CoordinatesError};

/// Maximum allowed accuracy in meters
const MAX_ACCURACY_METERS: f64 = 1000.0;

/// Maximum allowed speed in km/h
const MAX_SPEED_KMH: f64 = 200.0;

/// Maximum timestamp offset into the future (seconds)
const MAX_FUTURE_OFFSET_SECS: i64 = 60;

/// Maximum timestamp age (seconds)
const MAX_PAST_OFFSET_SECS: i64 = 300; // 5 minutes

/// Errors for location validation
#[derive(Debug, Error, Clone, PartialEq)]
pub enum LocationError {
    #[error("Coordinates error: {0}")]
    CoordinatesError(#[from] CoordinatesError),

    #[error("Invalid accuracy {0}: must be > 0 and <= {MAX_ACCURACY_METERS}")]
    InvalidAccuracy(f64),

    #[error("Invalid speed {0}: must be >= 0 and <= {MAX_SPEED_KMH}")]
    InvalidSpeed(f64),

    #[error("Invalid heading {0}: must be between 0 and 360")]
    InvalidHeading(f64),

    #[error("Timestamp is in the future")]
    TimestampInFuture,

    #[error("Timestamp is too old (> 5 minutes)")]
    TimestampTooOld,
}

/// GPS Location with full metadata
#[derive(Debug, Clone, PartialEq)]
pub struct Location {
    coordinates: Coordinates,
    accuracy: f64,
    timestamp: DateTime<Utc>,
    speed: Option<f64>,
    heading: Option<f64>,
}

impl Location {
    /// Create a new location with validation
    pub fn new(
        latitude: f64,
        longitude: f64,
        accuracy: f64,
        timestamp: DateTime<Utc>,
        speed: Option<f64>,
        heading: Option<f64>,
    ) -> Result<Self, LocationError> {
        let coordinates = Coordinates::new(latitude, longitude)?;

        // Validate accuracy
        if accuracy <= 0.0 || accuracy > MAX_ACCURACY_METERS {
            return Err(LocationError::InvalidAccuracy(accuracy));
        }

        // Validate speed if provided
        if let Some(s) = speed {
            if s < 0.0 || s > MAX_SPEED_KMH {
                return Err(LocationError::InvalidSpeed(s));
            }
        }

        // Validate heading if provided
        if let Some(h) = heading {
            if !(0.0..=360.0).contains(&h) {
                return Err(LocationError::InvalidHeading(h));
            }
        }

        // Validate timestamp
        let now = Utc::now();
        let diff = timestamp.signed_duration_since(now);
        if diff.num_seconds() > MAX_FUTURE_OFFSET_SECS {
            return Err(LocationError::TimestampInFuture);
        }
        if diff.num_seconds() < -MAX_PAST_OFFSET_SECS {
            return Err(LocationError::TimestampTooOld);
        }

        Ok(Self {
            coordinates,
            accuracy,
            timestamp,
            speed,
            heading,
        })
    }

    /// Create location without timestamp validation (for loading from DB)
    pub fn from_stored(
        latitude: f64,
        longitude: f64,
        accuracy: f64,
        timestamp: DateTime<Utc>,
        speed: Option<f64>,
        heading: Option<f64>,
    ) -> Result<Self, LocationError> {
        let coordinates = Coordinates::new(latitude, longitude)?;

        if accuracy <= 0.0 || accuracy > MAX_ACCURACY_METERS {
            return Err(LocationError::InvalidAccuracy(accuracy));
        }

        if let Some(s) = speed {
            if s < 0.0 || s > MAX_SPEED_KMH {
                return Err(LocationError::InvalidSpeed(s));
            }
        }

        if let Some(h) = heading {
            if !(0.0..=360.0).contains(&h) {
                return Err(LocationError::InvalidHeading(h));
            }
        }

        Ok(Self {
            coordinates,
            accuracy,
            timestamp,
            speed,
            heading,
        })
    }

    // === Getters ===

    pub fn coordinates(&self) -> &Coordinates {
        &self.coordinates
    }

    pub fn latitude(&self) -> f64 {
        self.coordinates.latitude()
    }

    pub fn longitude(&self) -> f64 {
        self.coordinates.longitude()
    }

    pub fn accuracy(&self) -> f64 {
        self.accuracy
    }

    pub fn timestamp(&self) -> DateTime<Utc> {
        self.timestamp
    }

    pub fn speed(&self) -> Option<f64> {
        self.speed
    }

    pub fn heading(&self) -> Option<f64> {
        self.heading
    }

    /// Calculate distance to another location in kilometers
    pub fn distance_to(&self, other: &Location) -> f64 {
        self.coordinates.distance_to(&other.coordinates)
    }

    /// Check if this location is within a certain radius (km) of another
    pub fn is_within(&self, other: &Location, radius_km: f64) -> bool {
        self.distance_to(other) <= radius_km
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn now() -> DateTime<Utc> {
        Utc::now()
    }

    #[test]
    fn test_valid_location() {
        let loc = Location::new(52.52, 13.405, 15.0, now(), Some(35.0), Some(180.0));
        assert!(loc.is_ok());
        let loc = loc.unwrap();
        assert_eq!(loc.latitude(), 52.52);
        assert_eq!(loc.longitude(), 13.405);
        assert_eq!(loc.accuracy(), 15.0);
        assert_eq!(loc.speed(), Some(35.0));
        assert_eq!(loc.heading(), Some(180.0));
    }

    #[test]
    fn test_location_without_optional_fields() {
        let loc = Location::new(52.52, 13.405, 20.0, now(), None, None);
        assert!(loc.is_ok());
        let loc = loc.unwrap();
        assert_eq!(loc.speed(), None);
        assert_eq!(loc.heading(), None);
    }

    #[test]
    fn test_invalid_accuracy() {
        assert!(matches!(
            Location::new(52.52, 13.405, 0.0, now(), None, None),
            Err(LocationError::InvalidAccuracy(_))
        ));
        assert!(matches!(
            Location::new(52.52, 13.405, 1001.0, now(), None, None),
            Err(LocationError::InvalidAccuracy(_))
        ));
    }

    #[test]
    fn test_invalid_speed() {
        assert!(matches!(
            Location::new(52.52, 13.405, 10.0, now(), Some(-1.0), None),
            Err(LocationError::InvalidSpeed(_))
        ));
        assert!(matches!(
            Location::new(52.52, 13.405, 10.0, now(), Some(201.0), None),
            Err(LocationError::InvalidSpeed(_))
        ));
    }

    #[test]
    fn test_invalid_heading() {
        assert!(matches!(
            Location::new(52.52, 13.405, 10.0, now(), None, Some(-1.0)),
            Err(LocationError::InvalidHeading(_))
        ));
        assert!(matches!(
            Location::new(52.52, 13.405, 10.0, now(), None, Some(361.0)),
            Err(LocationError::InvalidHeading(_))
        ));
    }

    #[test]
    fn test_distance_calculation() {
        let berlin = Location::new(52.52, 13.405, 10.0, now(), None, None).unwrap();
        let munich = Location::new(48.1351, 11.582, 10.0, now(), None, None).unwrap();
        let distance = berlin.distance_to(&munich);
        assert!((distance - 504.0).abs() < 10.0);
    }

    #[test]
    fn test_is_within() {
        let point1 = Location::new(52.52, 13.405, 10.0, now(), None, None).unwrap();
        let point2 = Location::new(52.521, 13.406, 10.0, now(), None, None).unwrap();
        assert!(point1.is_within(&point2, 1.0)); // Within 1km
        assert!(!point1.is_within(&point2, 0.001)); // Not within 1 meter
    }
}
