//! Get Courier Pool Query
//!
//! Data structure representing the query to retrieve couriers.

use crate::domain::model::courier::CourierStatus;
use crate::domain::model::vo::TransportType;

/// Filter criteria for courier pool
#[derive(Debug, Clone, Default)]
pub struct CourierFilter {
    /// Filter by status
    pub status: Option<CourierStatus>,
    /// Filter by work zone
    pub work_zone: Option<String>,
    /// Filter by transport type
    pub transport_type: Option<TransportType>,
    /// Only include couriers that can accept more packages
    pub available_only: bool,
}

impl CourierFilter {
    /// Create a filter for free couriers in a specific zone
    pub fn free_in_zone(zone: &str) -> Self {
        Self {
            status: Some(CourierStatus::Free),
            work_zone: Some(zone.to_string()),
            available_only: true,
            ..Default::default()
        }
    }

    /// Create a filter for all couriers in a zone
    pub fn in_zone(zone: &str) -> Self {
        Self {
            work_zone: Some(zone.to_string()),
            ..Default::default()
        }
    }
}

/// Query to get the courier pool
#[derive(Debug, Clone, Default)]
pub struct Query {
    /// Filter criteria
    pub filter: CourierFilter,
    /// Limit results
    pub limit: Option<u64>,
    /// Offset for pagination
    pub offset: Option<u64>,
}

impl Query {
    /// Create a new GetCourierPool query
    pub fn new(filter: CourierFilter) -> Self {
        Self { filter, limit: None, offset: None }
    }

    /// Create a query with no filters
    pub fn all() -> Self {
        Self::default()
    }

    /// Create a query for free couriers in a zone
    pub fn free_in_zone(zone: &str) -> Self {
        Self::new(CourierFilter::free_in_zone(zone))
    }

    /// Create a query for all couriers in a zone
    pub fn in_zone(zone: &str) -> Self {
        Self::new(CourierFilter::in_zone(zone))
    }

    /// Set pagination limit
    pub fn with_limit(mut self, limit: u64) -> Self {
        self.limit = Some(limit);
        self
    }

    /// Set pagination offset
    pub fn with_offset(mut self, offset: u64) -> Self {
        self.offset = Some(offset);
        self
    }
}

/// Type alias for backward compatibility
pub type Filter = CourierFilter;
