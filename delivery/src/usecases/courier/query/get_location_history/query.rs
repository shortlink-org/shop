//! Get Location History Query
//!
//! Data structure representing the query to get a courier's location history.

use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::domain::model::TimeRange;

/// Query to get a courier's location history
#[derive(Debug, Clone)]
pub struct Query {
    /// Courier ID
    pub courier_id: Uuid,
    /// Time range for the history
    pub time_range: TimeRange,
    /// Maximum number of entries to return (optional)
    pub limit: Option<u32>,
    /// Offset for pagination (optional)
    pub offset: Option<u32>,
}

impl Query {
    /// Create a new GetLocationHistory query
    pub fn new(courier_id: Uuid, time_range: TimeRange) -> Self {
        Self {
            courier_id,
            time_range,
            limit: None,
            offset: None,
        }
    }

    /// Create query with pagination
    pub fn paginated(
        courier_id: Uuid,
        time_range: TimeRange,
        limit: u32,
        offset: u32,
    ) -> Self {
        Self {
            courier_id,
            time_range,
            limit: Some(limit),
            offset: Some(offset),
        }
    }

    /// Create query for last N hours
    pub fn last_hours(courier_id: Uuid, hours: i64) -> Option<Self> {
        let end = Utc::now();
        let start = end - chrono::Duration::hours(hours);
        TimeRange::new(start, end).map(|time_range| Self::new(courier_id, time_range))
    }

    /// Create query for today
    pub fn today(courier_id: Uuid) -> Option<Self> {
        let now = Utc::now();
        let start = now.date_naive().and_hms_opt(0, 0, 0)?;
        let start = DateTime::from_naive_utc_and_offset(start, Utc);
        TimeRange::new(start, now).map(|time_range| Self::new(courier_id, time_range))
    }
}
