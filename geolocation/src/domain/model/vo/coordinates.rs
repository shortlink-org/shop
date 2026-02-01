//! Coordinates Value Object
//!
//! Represents validated geographic coordinates (latitude/longitude).

use thiserror::Error;

/// Errors for coordinate validation
#[derive(Debug, Error, Clone, PartialEq)]
pub enum CoordinatesError {
    #[error("Invalid latitude {0}: must be between -90 and 90")]
    InvalidLatitude(f64),
    #[error("Invalid longitude {0}: must be between -180 and 180")]
    InvalidLongitude(f64),
}

/// Geographic coordinates (latitude, longitude)
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct Coordinates {
    latitude: f64,
    longitude: f64,
}

impl Coordinates {
    /// Create new coordinates with validation
    pub fn new(latitude: f64, longitude: f64) -> Result<Self, CoordinatesError> {
        if !(-90.0..=90.0).contains(&latitude) {
            return Err(CoordinatesError::InvalidLatitude(latitude));
        }
        if !(-180.0..=180.0).contains(&longitude) {
            return Err(CoordinatesError::InvalidLongitude(longitude));
        }
        Ok(Self {
            latitude,
            longitude,
        })
    }

    /// Get latitude
    pub fn latitude(&self) -> f64 {
        self.latitude
    }

    /// Get longitude
    pub fn longitude(&self) -> f64 {
        self.longitude
    }

    /// Calculate Haversine distance to another point in kilometers
    pub fn distance_to(&self, other: &Coordinates) -> f64 {
        const EARTH_RADIUS_KM: f64 = 6371.0;

        let lat1_rad = self.latitude.to_radians();
        let lat2_rad = other.latitude.to_radians();
        let delta_lat = (other.latitude - self.latitude).to_radians();
        let delta_lon = (other.longitude - self.longitude).to_radians();

        let a = (delta_lat / 2.0).sin().powi(2)
            + lat1_rad.cos() * lat2_rad.cos() * (delta_lon / 2.0).sin().powi(2);
        let c = 2.0 * a.sqrt().asin();

        EARTH_RADIUS_KM * c
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_valid_coordinates() {
        let coords = Coordinates::new(52.52, 13.405);
        assert!(coords.is_ok());
        let coords = coords.unwrap();
        assert_eq!(coords.latitude(), 52.52);
        assert_eq!(coords.longitude(), 13.405);
    }

    #[test]
    fn test_invalid_latitude() {
        assert!(matches!(
            Coordinates::new(91.0, 0.0),
            Err(CoordinatesError::InvalidLatitude(_))
        ));
        assert!(matches!(
            Coordinates::new(-91.0, 0.0),
            Err(CoordinatesError::InvalidLatitude(_))
        ));
    }

    #[test]
    fn test_invalid_longitude() {
        assert!(matches!(
            Coordinates::new(0.0, 181.0),
            Err(CoordinatesError::InvalidLongitude(_))
        ));
        assert!(matches!(
            Coordinates::new(0.0, -181.0),
            Err(CoordinatesError::InvalidLongitude(_))
        ));
    }

    #[test]
    fn test_haversine_distance() {
        // Berlin to Munich ~500km
        let berlin = Coordinates::new(52.52, 13.405).unwrap();
        let munich = Coordinates::new(48.1351, 11.582).unwrap();
        let distance = berlin.distance_to(&munich);
        assert!((distance - 504.0).abs() < 10.0); // Allow 10km tolerance
    }

    #[test]
    fn test_same_point_distance() {
        let point = Coordinates::new(52.52, 13.405).unwrap();
        assert_eq!(point.distance_to(&point), 0.0);
    }
}
