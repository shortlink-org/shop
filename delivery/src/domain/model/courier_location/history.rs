//! Location History Entry
//!
//! Represents a historical record of a courier's location.
//! Used for tracking courier movement over time.

use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::domain::model::vo::location::Location;

/// Location history entry for audit and analytics
#[derive(Debug, Clone, PartialEq)]
pub struct LocationHistoryEntry {
    /// Unique ID of this history entry
    id: Uuid,
    /// Courier ID
    courier_id: Uuid,
    /// GPS location
    location: Location,
    /// Timestamp when location was recorded
    timestamp: DateTime<Utc>,
    /// Speed in km/h (optional)
    speed: Option<f64>,
    /// Heading in degrees 0-360 (optional)
    heading: Option<f64>,
    /// When this entry was stored in the database
    created_at: DateTime<Utc>,
}

impl LocationHistoryEntry {
    /// Create a new location history entry
    pub fn new(
        courier_id: Uuid,
        location: Location,
        timestamp: DateTime<Utc>,
        speed: Option<f64>,
        heading: Option<f64>,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            courier_id,
            location,
            timestamp,
            speed,
            heading,
            created_at: Utc::now(),
        }
    }

    /// Reconstitute from storage
    pub fn reconstitute(
        id: Uuid,
        courier_id: Uuid,
        location: Location,
        timestamp: DateTime<Utc>,
        speed: Option<f64>,
        heading: Option<f64>,
        created_at: DateTime<Utc>,
    ) -> Self {
        Self {
            id,
            courier_id,
            location,
            timestamp,
            speed,
            heading,
            created_at,
        }
    }

    // === Getters ===

    /// Get entry ID
    pub fn id(&self) -> Uuid {
        self.id
    }

    /// Get courier ID
    pub fn courier_id(&self) -> Uuid {
        self.courier_id
    }

    /// Get location
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

    /// Get accuracy
    pub fn accuracy(&self) -> f64 {
        self.location.accuracy()
    }

    /// Get timestamp
    pub fn timestamp(&self) -> DateTime<Utc> {
        self.timestamp
    }

    /// Get speed
    pub fn speed(&self) -> Option<f64> {
        self.speed
    }

    /// Get heading
    pub fn heading(&self) -> Option<f64> {
        self.heading
    }

    /// Get created_at
    pub fn created_at(&self) -> DateTime<Utc> {
        self.created_at
    }
}

/// Time range for querying location history
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct TimeRange {
    start: DateTime<Utc>,
    end: DateTime<Utc>,
}

impl TimeRange {
    /// Create a new time range
    ///
    /// # Arguments
    ///
    /// * `start` - Start of the range (inclusive)
    /// * `end` - End of the range (inclusive)
    ///
    /// # Errors
    ///
    /// Returns None if start is after end
    pub fn new(start: DateTime<Utc>, end: DateTime<Utc>) -> Option<Self> {
        if start > end {
            return None;
        }
        Some(Self { start, end })
    }

    /// Get start of range
    pub fn start(&self) -> DateTime<Utc> {
        self.start
    }

    /// Get end of range
    pub fn end(&self) -> DateTime<Utc> {
        self.end
    }

    /// Check if a timestamp is within this range
    pub fn contains(&self, timestamp: DateTime<Utc>) -> bool {
        timestamp >= self.start && timestamp <= self.end
    }

    /// Get duration of the range in seconds
    pub fn duration_secs(&self) -> i64 {
        (self.end - self.start).num_seconds()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_location() -> Location {
        Location::new(52.52, 13.405, 10.0).unwrap()
    }

    #[test]
    fn test_location_history_entry_creation() {
        let courier_id = Uuid::new_v4();
        let location = create_location();
        let timestamp = Utc::now();

        let entry = LocationHistoryEntry::new(
            courier_id,
            location,
            timestamp,
            Some(35.0),
            Some(180.0),
        );

        assert_eq!(entry.courier_id(), courier_id);
        assert_eq!(entry.latitude(), 52.52);
        assert_eq!(entry.longitude(), 13.405);
        assert_eq!(entry.speed(), Some(35.0));
        assert_eq!(entry.heading(), Some(180.0));
        assert!(entry.id() != Uuid::nil());
    }

    #[test]
    fn test_time_range() {
        let start = Utc::now() - chrono::Duration::hours(1);
        let end = Utc::now();

        let range = TimeRange::new(start, end);
        assert!(range.is_some());
        let range = range.unwrap();

        assert_eq!(range.start(), start);
        assert_eq!(range.end(), end);
        assert!(range.duration_secs() > 0);
    }

    #[test]
    fn test_time_range_invalid() {
        let start = Utc::now();
        let end = Utc::now() - chrono::Duration::hours(1);

        let range = TimeRange::new(start, end);
        assert!(range.is_none());
    }

    #[test]
    fn test_time_range_contains() {
        let start = Utc::now() - chrono::Duration::hours(2);
        let end = Utc::now();
        let range = TimeRange::new(start, end).unwrap();

        let middle = Utc::now() - chrono::Duration::hours(1);
        let before = Utc::now() - chrono::Duration::hours(3);
        let after = Utc::now() + chrono::Duration::hours(1);

        assert!(range.contains(middle));
        assert!(range.contains(start)); // inclusive
        assert!(range.contains(end)); // inclusive
        assert!(!range.contains(before));
        assert!(!range.contains(after));
    }
}
