use std::fmt;

/// Constants for location validation bounds
pub const MIN_LATITUDE: f64 = -90.0;
pub const MAX_LATITUDE: f64 = 90.0;
pub const MIN_LONGITUDE: f64 = -180.0;
pub const MAX_LONGITUDE: f64 = 180.0;

/// Location represents a GPS location as a value object.
/// 
/// A value object is immutable and defined by its attributes.
/// Two locations are considered equal if they have the same latitude and longitude.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct Location {
    /// Latitude in degrees (MIN_LATITUDE to MAX_LATITUDE)
    latitude: f64,
    /// Longitude in degrees (MIN_LONGITUDE to MAX_LONGITUDE)
    longitude: f64,
    /// Accuracy in meters (must be positive)
    accuracy: f64,
}

#[derive(Debug, Clone, PartialEq)]
pub enum LocationError {
    /// Latitude is out of valid range (MIN_LATITUDE to MAX_LATITUDE)
    InvalidLatitude(f64),
    /// Longitude is out of valid range (MIN_LONGITUDE to MAX_LONGITUDE)
    InvalidLongitude(f64),
    /// Accuracy must be positive
    InvalidAccuracy(f64),
}

impl fmt::Display for LocationError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            LocationError::InvalidLatitude(lat) => {
                write!(f, "Invalid latitude: {}. Must be between {} and {}", lat, MIN_LATITUDE, MAX_LATITUDE)
            }
            LocationError::InvalidLongitude(lon) => {
                write!(f, "Invalid longitude: {}. Must be between {} and {}", lon, MIN_LONGITUDE, MAX_LONGITUDE)
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
    /// * `latitude` - Latitude in degrees (MIN_LATITUDE to MAX_LATITUDE)
    /// * `longitude` - Longitude in degrees (MIN_LONGITUDE to MAX_LONGITUDE)
    /// * `accuracy` - Accuracy in meters (must be positive)
    ///
    /// # Errors
    ///
    /// Returns `LocationError` if any of the values are invalid.
    ///
    /// # Examples
    ///
    /// ```
    /// use delivery::domain::model::vo::location::{Location, LocationError};
    ///
    /// let location = Location::new(55.7558, 37.6173, 10.0)?;
    /// # Ok::<(), LocationError>(())
    /// ```
    pub fn new(latitude: f64, longitude: f64, accuracy: f64) -> Result<Self, LocationError> {
        if !(MIN_LATITUDE..=MAX_LATITUDE).contains(&latitude) {
            return Err(LocationError::InvalidLatitude(latitude));
        }

        if !(MIN_LONGITUDE..=MAX_LONGITUDE).contains(&longitude) {
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
    /// For valid `Location` values (constructed via `new`), the result is always
    /// finite and not NaN. Callers may use `total_cmp` for stable sorting.
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
    /// use delivery::domain::model::vo::location::Location;
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
        let location = Location::new(MAX_LATITUDE + 1.0, 37.6173, 10.0);
        assert!(matches!(location, Err(LocationError::InvalidLatitude(_))));

        let location = Location::new(MIN_LATITUDE - 1.0, 37.6173, 10.0);
        assert!(matches!(location, Err(LocationError::InvalidLatitude(_))));
    }

    #[test]
    fn test_location_creation_invalid_longitude() {
        let location = Location::new(55.7558, MAX_LONGITUDE + 1.0, 10.0);
        assert!(matches!(location, Err(LocationError::InvalidLongitude(_))));

        let location = Location::new(55.7558, MIN_LONGITUDE - 1.0, 10.0);
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
        // Test boundary values using constants
        assert!(Location::new(MAX_LATITUDE, 0.0, 1.0).is_ok());
        assert!(Location::new(MIN_LATITUDE, 0.0, 1.0).is_ok());
        assert!(Location::new(0.0, MAX_LONGITUDE, 1.0).is_ok());
        assert!(Location::new(0.0, MIN_LONGITUDE, 1.0).is_ok());
    }

    #[test]
    fn test_special_coordinates() {
        // Equator (latitude = 0)
        assert!(Location::new(0.0, 0.0, 1.0).is_ok());
        
        // Prime meridian (longitude = 0)
        assert!(Location::new(0.0, 0.0, 1.0).is_ok());
        
        // North pole
        assert!(Location::new(90.0, 0.0, 1.0).is_ok());
        
        // South pole
        assert!(Location::new(-90.0, 0.0, 1.0).is_ok());
        
        // Date line (longitude = 180/-180)
        assert!(Location::new(0.0, 180.0, 1.0).is_ok());
        assert!(Location::new(0.0, -180.0, 1.0).is_ok());
    }

    #[test]
    fn test_multiple_invalid_parameters() {
        // Multiple invalid parameters - should return first error encountered (latitude)
        let location = Location::new(MAX_LATITUDE + 1.0, MAX_LONGITUDE + 1.0, -1.0);
        assert!(matches!(location, Err(LocationError::InvalidLatitude(_))));
    }

    #[test]
    fn test_equality_includes_accuracy() {
        // Two locations with same lat/lon but different accuracy should NOT be equal
        // (all fields matter for value object equality)
        let loc1 = Location::new(55.7558, 37.6173, 10.0).unwrap();
        let loc2 = Location::new(55.7558, 37.6173, 100.0).unwrap();
        
        assert_ne!(loc1, loc2);
        assert_eq!(loc1.latitude(), loc2.latitude());
        assert_eq!(loc1.longitude(), loc2.longitude());
        assert_ne!(loc1.accuracy(), loc2.accuracy());
    }

    #[test]
    fn test_distance_zero() {
        // Distance to self should be zero
        let loc = Location::new(55.7558, 37.6173, 10.0).unwrap();
        assert_eq!(loc.distance_to(&loc), 0.0);
    }

    #[test]
    fn test_distance_same_coordinates() {
        // Distance between two locations with same coordinates should be zero
        let loc1 = Location::new(55.7558, 37.6173, 10.0).unwrap();
        let loc2 = Location::new(55.7558, 37.6173, 20.0).unwrap();
        assert_eq!(loc1.distance_to(&loc2), 0.0);
    }

    #[test]
    fn test_distance_antipodal_points() {
        // Distance between antipodal points (opposite sides of Earth)
        let loc1 = Location::new(0.0, 0.0, 10.0).unwrap();
        let loc2 = Location::new(0.0, 180.0, 10.0).unwrap();
        let distance = loc1.distance_to(&loc2);
        
        // Should be approximately half the Earth's circumference (~20,000 km)
        assert!((19000.0..21000.0).contains(&distance));
    }

    #[test]
    fn test_distance_symmetry() {
        // Distance from A to B should equal distance from B to A
        let moscow = Location::new(55.7558, 37.6173, 10.0).unwrap();
        let spb = Location::new(59.9343, 30.3351, 10.0).unwrap();
        
        assert_eq!(moscow.distance_to(&spb), spb.distance_to(&moscow));
    }

    #[test]
    fn test_is_within_distance_zero_distance() {
        // Location should be within any positive distance from itself
        let loc = Location::new(55.7558, 37.6173, 10.0).unwrap();
        assert!(loc.is_within_distance(&loc, 0.0));
        assert!(loc.is_within_distance(&loc, 1.0));
        assert!(loc.is_within_distance(&loc, 1000.0));
    }

    #[test]
    fn test_is_within_distance_exact_boundary() {
        // Test exact boundary case
        let loc1 = Location::new(55.7558, 37.6173, 10.0).unwrap();
        let loc2 = Location::new(55.7560, 37.6175, 10.0).unwrap();
        
        let distance = loc1.distance_to(&loc2);
        assert!(loc1.is_within_distance(&loc2, distance));
        assert!(loc1.is_within_distance(&loc2, distance + 0.1));
        assert!(!loc1.is_within_distance(&loc2, distance - 0.1));
    }

    #[test]
    fn test_error_display_invalid_latitude() {
        let err = LocationError::InvalidLatitude(95.0);
        let display = format!("{}", err);
        assert!(display.contains("Invalid latitude"));
        assert!(display.contains("95"));
        assert!(display.contains(&MIN_LATITUDE.to_string()));
        assert!(display.contains(&MAX_LATITUDE.to_string()));
    }

    #[test]
    fn test_error_display_invalid_longitude() {
        let err = LocationError::InvalidLongitude(185.0);
        let display = format!("{}", err);
        assert!(display.contains("Invalid longitude"));
        assert!(display.contains("185"));
        assert!(display.contains(&MIN_LONGITUDE.to_string()));
        assert!(display.contains(&MAX_LONGITUDE.to_string()));
    }

    #[test]
    fn test_error_display_invalid_accuracy() {
        let err = LocationError::InvalidAccuracy(-5.0);
        let display = format!("{}", err);
        assert!(display.contains("Invalid accuracy"));
        assert!(display.contains("-5"));
        assert!(display.contains("positive"));
    }

    #[test]
    fn test_clone() {
        let loc1 = Location::new(55.7558, 37.6173, 10.0).unwrap();
        let loc2 = loc1.clone();
        
        assert_eq!(loc1, loc2);
        assert_eq!(loc1.latitude(), loc2.latitude());
        assert_eq!(loc1.longitude(), loc2.longitude());
        assert_eq!(loc1.accuracy(), loc2.accuracy());
    }

    #[test]
    fn test_copy() {
        // Test that Location implements Copy
        let loc1 = Location::new(55.7558, 37.6173, 10.0).unwrap();
        let loc2 = loc1; // This should copy, not move
        
        // Both should be valid
        assert_eq!(loc1.latitude(), loc2.latitude());
        assert_eq!(loc1.longitude(), loc2.longitude());
    }

    #[test]
    fn test_debug() {
        let loc = Location::new(55.7558, 37.6173, 10.0).unwrap();
        let debug_str = format!("{:?}", loc);
        assert!(debug_str.contains("Location"));
    }

    #[test]
    fn test_accuracy_edge_cases() {
        // Very small positive accuracy should be valid
        assert!(Location::new(0.0, 0.0, 0.0001).is_ok());
        
        // Large accuracy should be valid
        assert!(Location::new(0.0, 0.0, 1_000_000.0).is_ok());
        
        // Negative accuracy should be invalid
        assert!(matches!(
            Location::new(0.0, 0.0, -0.0001),
            Err(LocationError::InvalidAccuracy(_))
        ));
    }

    #[test]
    fn test_distance_precision() {
        // Test that distance calculation maintains reasonable precision
        let loc1 = Location::new(55.7558, 37.6173, 10.0).unwrap();
        let loc2 = Location::new(55.7558001, 37.6173001, 10.0).unwrap();
        
        let distance = loc1.distance_to(&loc2);
        // Should be very small but non-zero
        assert!(distance > 0.0);
        assert!(distance < 0.1); // Less than 100 meters
    }

    #[test]
    fn test_getters() {
        let latitude = 55.7558;
        let longitude = 37.6173;
        let accuracy = 10.0;
        
        let loc = Location::new(latitude, longitude, accuracy).unwrap();
        
        assert_eq!(loc.latitude(), latitude);
        assert_eq!(loc.longitude(), longitude);
        assert_eq!(loc.accuracy(), accuracy);
    }
}

