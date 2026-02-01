//! Courier Location Entity
//!
//! Represents the current location of a courier.

use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::domain::model::vo::Location;

/// Courier's current location (aggregate root)
#[derive(Debug, Clone)]
pub struct CourierLocation {
    courier_id: Uuid,
    location: Location,
    updated_at: DateTime<Utc>,
}

impl CourierLocation {
    /// Create a new courier location
    pub fn new(courier_id: Uuid, location: Location) -> Self {
        Self {
            courier_id,
            location,
            updated_at: Utc::now(),
        }
    }

    /// Reconstitute from storage
    pub fn reconstitute(
        courier_id: Uuid,
        location: Location,
        updated_at: DateTime<Utc>,
    ) -> Self {
        Self {
            courier_id,
            location,
            updated_at,
        }
    }

    /// Update location
    pub fn update(&mut self, location: Location) {
        self.location = location;
        self.updated_at = Utc::now();
    }

    // === Getters ===

    pub fn courier_id(&self) -> Uuid {
        self.courier_id
    }

    pub fn location(&self) -> &Location {
        &self.location
    }

    pub fn latitude(&self) -> f64 {
        self.location.latitude()
    }

    pub fn longitude(&self) -> f64 {
        self.location.longitude()
    }

    pub fn accuracy(&self) -> f64 {
        self.location.accuracy()
    }

    pub fn speed(&self) -> Option<f64> {
        self.location.speed()
    }

    pub fn heading(&self) -> Option<f64> {
        self.location.heading()
    }

    pub fn timestamp(&self) -> DateTime<Utc> {
        self.location.timestamp()
    }

    pub fn updated_at(&self) -> DateTime<Utc> {
        self.updated_at
    }

    /// Calculate distance to another courier's location
    pub fn distance_to(&self, other: &CourierLocation) -> f64 {
        self.location.distance_to(&other.location)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_test_location() -> Location {
        Location::from_stored(52.52, 13.405, 15.0, Utc::now(), Some(35.0), Some(180.0)).unwrap()
    }

    #[test]
    fn test_courier_location_creation() {
        let courier_id = Uuid::new_v4();
        let location = create_test_location();
        let courier_loc = CourierLocation::new(courier_id, location);

        assert_eq!(courier_loc.courier_id(), courier_id);
        assert_eq!(courier_loc.latitude(), 52.52);
        assert_eq!(courier_loc.longitude(), 13.405);
    }

    #[test]
    fn test_courier_location_update() {
        let courier_id = Uuid::new_v4();
        let location1 = create_test_location();
        let mut courier_loc = CourierLocation::new(courier_id, location1);

        let location2 = Location::from_stored(48.1351, 11.582, 20.0, Utc::now(), None, None).unwrap();
        courier_loc.update(location2);

        assert_eq!(courier_loc.latitude(), 48.1351);
        assert_eq!(courier_loc.longitude(), 11.582);
    }
}
