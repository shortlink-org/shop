//! Get Location History Query
//!
//! Query to retrieve location history for a courier.

use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::domain::model::TimeRange;

/// Default limit for history entries
pub const DEFAULT_LIMIT: u32 = 100;

/// Maximum limit for history entries
pub const MAX_LIMIT: u32 = 1000;

/// Query to get courier's location history
#[derive(Debug, Clone)]
pub struct Query {
    /// Courier ID
    pub courier_id: Uuid,
    /// Time range for the history
    pub time_range: TimeRange,
    /// Maximum number of entries to return
    pub limit: Option<u32>,
}

impl Query {
    /// Create a new query
    pub fn new(courier_id: Uuid, start: DateTime<Utc>, end: DateTime<Utc>, limit: Option<u32>) -> Option<Self> {
        let time_range = TimeRange::new(start, end)?;
        Some(Self {
            courier_id,
            time_range,
            limit,
        })
    }

    /// Get effective limit
    pub fn effective_limit(&self) -> u32 {
        self.limit.unwrap_or(DEFAULT_LIMIT).min(MAX_LIMIT)
    }
}
