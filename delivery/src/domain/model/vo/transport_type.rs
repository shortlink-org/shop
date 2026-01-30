//! Transport Type Value Object
//!
//! Represents the type of transport a courier uses for deliveries.
//! This is a value object - immutable and defined entirely by its value.

use std::fmt;

/// Transport type for courier
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum TransportType {
    Walking,
    Bicycle,
    Motorcycle,
    Car,
}

impl TransportType {
    /// Get average speed in km/h for this transport type
    pub fn average_speed_kmh(&self) -> f64 {
        match self {
            TransportType::Walking => 5.0,
            TransportType::Bicycle => 15.0,
            TransportType::Motorcycle => 40.0,
            TransportType::Car => 30.0, // Lower due to traffic
        }
    }

    /// Get maximum recommended distance in km
    pub fn max_recommended_distance_km(&self) -> f64 {
        match self {
            TransportType::Walking => 3.0,
            TransportType::Bicycle => 10.0,
            TransportType::Motorcycle => 30.0,
            TransportType::Car => 50.0,
        }
    }

    /// Calculate travel time in minutes for a given distance
    pub fn calculate_travel_time_minutes(&self, distance_km: f64) -> f64 {
        (distance_km / self.average_speed_kmh()) * 60.0
    }
}

impl fmt::Display for TransportType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            TransportType::Walking => write!(f, "Walking"),
            TransportType::Bicycle => write!(f, "Bicycle"),
            TransportType::Motorcycle => write!(f, "Motorcycle"),
            TransportType::Car => write!(f, "Car"),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_transport_speeds() {
        assert!(TransportType::Car.average_speed_kmh() > TransportType::Walking.average_speed_kmh());
        assert!(
            TransportType::Motorcycle.average_speed_kmh() > TransportType::Bicycle.average_speed_kmh()
        );
    }

    #[test]
    fn test_max_recommended_distance() {
        assert!(
            TransportType::Car.max_recommended_distance_km()
                > TransportType::Walking.max_recommended_distance_km()
        );
        assert!(
            TransportType::Motorcycle.max_recommended_distance_km()
                > TransportType::Bicycle.max_recommended_distance_km()
        );
    }

    #[test]
    fn test_display() {
        assert_eq!(format!("{}", TransportType::Walking), "Walking");
        assert_eq!(format!("{}", TransportType::Bicycle), "Bicycle");
        assert_eq!(format!("{}", TransportType::Motorcycle), "Motorcycle");
        assert_eq!(format!("{}", TransportType::Car), "Car");
    }

    #[test]
    fn test_equality() {
        assert_eq!(TransportType::Car, TransportType::Car);
        assert_ne!(TransportType::Car, TransportType::Bicycle);
    }

    #[test]
    fn test_clone_and_copy() {
        let t1 = TransportType::Motorcycle;
        let t2 = t1; // Copy
        let t3 = t1.clone();

        assert_eq!(t1, t2);
        assert_eq!(t1, t3);
    }

    #[test]
    fn test_calculate_travel_time_minutes() {
        // Car: 30 km/h, 15 km distance = 30 minutes
        assert_eq!(TransportType::Car.calculate_travel_time_minutes(15.0), 30.0);

        // Walking: 5 km/h, 5 km distance = 60 minutes
        assert_eq!(TransportType::Walking.calculate_travel_time_minutes(5.0), 60.0);

        // Bicycle: 15 km/h, 15 km distance = 60 minutes
        assert_eq!(TransportType::Bicycle.calculate_travel_time_minutes(15.0), 60.0);

        // Motorcycle: 40 km/h, 20 km distance = 30 minutes
        assert_eq!(TransportType::Motorcycle.calculate_travel_time_minutes(20.0), 30.0);

        // Zero distance = zero time
        assert_eq!(TransportType::Car.calculate_travel_time_minutes(0.0), 0.0);
    }
}
