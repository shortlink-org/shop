//! Location History Entry Entity
//!
//! Represents a single entry in the courier's location history.

use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::domain::model::vo::Location;

/// A single location history entry
#[derive(Debug, Clone)]
pub struct LocationHistoryEntry {
    id: Uuid,
    courier_id: Uuid,
    location: Location,
    recorded_at: DateTime<Utc>,
}

impl LocationHistoryEntry {
    /// Create a new history entry
    pub fn new(courier_id: Uuid, location: Location) -> Self {
        Self {
            id: Uuid::new_v4(),
            courier_id,
            location,
            recorded_at: Utc::now(),
        }
    }

    /// Reconstitute from storage
    pub fn reconstitute(
        id: Uuid,
        courier_id: Uuid,
        location: Location,
        recorded_at: DateTime<Utc>,
    ) -> Self {
        Self {
            id,
            courier_id,
            location,
            recorded_at,
        }
    }

    // === Getters ===

    pub fn id(&self) -> Uuid {
        self.id
    }

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

    pub fn recorded_at(&self) -> DateTime<Utc> {
        self.recorded_at
    }
}

/// Time range for querying location history
#[derive(Debug, Clone)]
pub struct TimeRange {
    start: DateTime<Utc>,
    end: DateTime<Utc>,
}

impl TimeRange {
    /// Create a new time range
    pub fn new(start: DateTime<Utc>, end: DateTime<Utc>) -> Option<Self> {
        if start >= end {
            return None;
        }
        Some(Self { start, end })
    }

    pub fn start(&self) -> DateTime<Utc> {
        self.start
    }

    pub fn end(&self) -> DateTime<Utc> {
        self.end
    }

    /// Check if a timestamp is within this range
    pub fn contains(&self, timestamp: DateTime<Utc>) -> bool {
        timestamp >= self.start && timestamp <= self.end
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::Duration;

    fn create_test_location() -> Location {
        Location::from_stored(52.52, 13.405, 15.0, Utc::now(), Some(35.0), Some(180.0)).unwrap()
    }

    #[test]
    fn test_history_entry_creation() {
        let courier_id = Uuid::new_v4();
        let location = create_test_location();
        let entry = LocationHistoryEntry::new(courier_id, location);

        assert_eq!(entry.courier_id(), courier_id);
        assert_eq!(entry.latitude(), 52.52);
    }

    #[test]
    fn test_time_range() {
        let now = Utc::now();
        let range = TimeRange::new(now - Duration::hours(1), now);
        assert!(range.is_some());

        let range = range.unwrap();
        assert!(range.contains(now - Duration::minutes(30)));
        assert!(!range.contains(now - Duration::hours(2)));
    }

    #[test]
    fn test_invalid_time_range() {
        let now = Utc::now();
        let range = TimeRange::new(now, now - Duration::hours(1));
        assert!(range.is_none());
    }
}
