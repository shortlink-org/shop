use std::fmt;

/// Location represents a GPS location as a value object.
/// 
/// A value object is immutable and defined by its attributes.
/// Two locations are considered equal if they have the same latitude and longitude.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct Location {
    /// Latitude in degrees (-90.0 to 90.0)
    latitude: f64,
    /// Longitude in degrees (-180.0 to 180.0)
    longitude: f64,
    /// Accuracy in meters (must be positive)
    accuracy: f64,
}

#[derive(Debug, Clone, PartialEq)]
pub enum LocationError {
    /// Latitude is out of valid range (-90.0 to 90.0)
    InvalidLatitude(f64),
    /// Longitude is out of valid range (-180.0 to 180.0)
    InvalidLongitude(f64),
    /// Accuracy must be positive
    InvalidAccuracy(f64),
}

impl fmt::Display for LocationError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            LocationError::InvalidLatitude(lat) => {
                write!(f, "Invalid latitude: {}. Must be between -90.0 and 90.0", lat)
            }
            LocationError::InvalidLongitude(lon) => {
                write!(f, "Invalid longitude: {}. Must be between -180.0 and 180.0", lon)
            }
            LocationError::InvalidAccuracy(acc) => {
                write!(f, "Invalid accuracy: {}. Must be positive", acc)
            }
        }
    }
}

impl std::error::Error for LocationError {}

impl Location {
    /// Creates a new Location value object with validation.
    ///
    /// # Arguments
    ///
    /// * `latitude` - Latitude in degrees (-90.0 to 90.0)
    /// * `longitude` - Longitude in degrees (-180.0 to 180.0)
    /// * `accuracy` - Accuracy in meters (must be positive)
    ///
    /// # Errors
    ///
    /// Returns `LocationError` if any of the values are invalid.
    ///
    /// # Examples
    ///
    /// ```
    /// use crate::domain::value_objects::location::{Location, LocationError};
    ///
    /// let location = Location::new(55.7558, 37.6173, 10.0)?;
    /// # Ok::<(), LocationError>(())
    /// ```
    pub fn new(latitude: f64, longitude: f64, accuracy: f64) -> Result<Self, LocationError> {
        if !(-90.0..=90.0).contains(&latitude) {
            return Err(LocationError::InvalidLatitude(latitude));
        }

        if !(-180.0..=180.0).contains(&longitude) {
            return Err(LocationError::InvalidLongitude(longitude));
        }

        if accuracy <= 0.0 {
            return Err(LocationError::InvalidAccuracy(accuracy));
        }

        Ok(Self {
            latitude,
            longitude,
            accuracy,
        })
    }

    /// Returns the latitude in degrees.
    #[inline]
    pub fn latitude(&self) -> f64 {
        self.latitude
    }

    /// Returns the longitude in degrees.
    #[inline]
    pub fn longitude(&self) -> f64 {
        self.longitude
    }

    /// Returns the accuracy in meters.
    #[inline]
    pub fn accuracy(&self) -> f64 {
        self.accuracy
    }

    /// Calculates the distance to another location using the Haversine formula.
    ///
    /// # Arguments
    ///
    /// * `other` - The other location to calculate distance to
    ///
    /// # Returns
    ///
    /// Distance in kilometers.
    ///
    /// # Examples
    ///
    /// ```
    /// use crate::domain::value_objects::location::Location;
    ///
    /// let moscow = Location::new(55.7558, 37.6173, 10.0).unwrap();
    /// let spb = Location::new(59.9343, 30.3351, 10.0).unwrap();
    /// let distance = moscow.distance_to(&spb);
    /// println!("Distance: {:.2} km", distance);
    /// ```
    pub fn distance_to(&self, other: &Location) -> f64 {
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

    /// Checks if this location is within a certain distance of another location.
    ///
    /// # Arguments
    ///
    /// * `other` - The other location to check
    /// * `max_distance_km` - Maximum distance in kilometers
    ///
    /// # Returns
    ///
    /// `true` if the distance is less than or equal to `max_distance_km`, `false` otherwise.
    pub fn is_within_distance(&self, other: &Location, max_distance_km: f64) -> bool {
        self.distance_to(other) <= max_distance_km
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_location_creation_valid() {
        let location = Location::new(55.7558, 37.6173, 10.0);
        assert!(location.is_ok());
        let loc = location.unwrap();
        assert_eq!(loc.latitude(), 55.7558);
        assert_eq!(loc.longitude(), 37.6173);
        assert_eq!(loc.accuracy(), 10.0);
    }

    #[test]
    fn test_location_creation_invalid_latitude() {
        let location = Location::new(91.0, 37.6173, 10.0);
        assert!(matches!(location, Err(LocationError::InvalidLatitude(_))));

        let location = Location::new(-91.0, 37.6173, 10.0);
        assert!(matches!(location, Err(LocationError::InvalidLatitude(_))));
    }

    #[test]
    fn test_location_creation_invalid_longitude() {
        let location = Location::new(55.7558, 181.0, 10.0);
        assert!(matches!(location, Err(LocationError::InvalidLongitude(_))));

        let location = Location::new(55.7558, -181.0, 10.0);
        assert!(matches!(location, Err(LocationError::InvalidLongitude(_))));
    }

    #[test]
    fn test_location_creation_invalid_accuracy() {
        let location = Location::new(55.7558, 37.6173, -1.0);
        assert!(matches!(location, Err(LocationError::InvalidAccuracy(_))));

        let location = Location::new(55.7558, 37.6173, 0.0);
        assert!(matches!(location, Err(LocationError::InvalidAccuracy(_))));
    }

    #[test]
    fn test_location_equality() {
        let loc1 = Location::new(55.7558, 37.6173, 10.0).unwrap();
        let loc2 = Location::new(55.7558, 37.6173, 10.0).unwrap();
        let loc3 = Location::new(59.9343, 30.3351, 10.0).unwrap();

        assert_eq!(loc1, loc2);
        assert_ne!(loc1, loc3);
    }

    #[test]
    fn test_distance_calculation() {
        // Moscow coordinates
        let moscow = Location::new(55.7558, 37.6173, 10.0).unwrap();
        // St. Petersburg coordinates
        let spb = Location::new(59.9343, 30.3351, 10.0).unwrap();

        let distance = moscow.distance_to(&spb);
        // Distance between Moscow and St. Petersburg is approximately 635 km
        assert!((600.0..700.0).contains(&distance));
    }

    #[test]
    fn test_is_within_distance() {
        let loc1 = Location::new(55.7558, 37.6173, 10.0).unwrap();
        let loc2 = Location::new(55.7560, 37.6175, 10.0).unwrap(); // Very close
        let loc3 = Location::new(59.9343, 30.3351, 10.0).unwrap(); // Far away

        assert!(loc1.is_within_distance(&loc2, 1.0)); // Within 1 km
        assert!(!loc1.is_within_distance(&loc3, 100.0)); // Not within 100 km
        assert!(loc1.is_within_distance(&loc3, 700.0)); // Within 700 km
    }

    #[test]
    fn test_boundary_values() {
        // Test boundary values
        assert!(Location::new(90.0, 0.0, 1.0).is_ok());
        assert!(Location::new(-90.0, 0.0, 1.0).is_ok());
        assert!(Location::new(0.0, 180.0, 1.0).is_ok());
        assert!(Location::new(0.0, -180.0, 1.0).is_ok());
    }
}

