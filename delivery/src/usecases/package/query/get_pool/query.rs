//! Get Package Pool Query
//!
//! Data structure representing the query to retrieve packages.

use uuid::Uuid;

use crate::domain::model::package::PackageStatus;

/// Filter criteria for package pool
#[derive(Debug, Clone, Default)]
pub struct PackageFilter {
    /// Filter by status
    pub status: Option<PackageStatus>,
    /// Filter by delivery zone
    pub zone: Option<String>,
    /// Filter by assigned courier
    pub courier_id: Option<Uuid>,
    /// Filter by priority (minimum)
    pub min_priority: Option<u8>,
    /// Only include unassigned packages
    pub unassigned_only: bool,
}

impl PackageFilter {
    /// Create a filter for packages in pool (waiting for assignment)
    pub fn in_pool() -> Self {
        Self {
            status: Some(PackageStatus::InPool),
            unassigned_only: true,
            ..Default::default()
        }
    }

    /// Create a filter for packages in a specific zone
    pub fn in_zone(zone: &str) -> Self {
        Self {
            zone: Some(zone.to_string()),
            ..Default::default()
        }
    }

    /// Create a filter for packages assigned to a courier
    pub fn by_courier(courier_id: Uuid) -> Self {
        Self {
            courier_id: Some(courier_id),
            ..Default::default()
        }
    }

    /// Create a filter for high priority packages
    pub fn high_priority() -> Self {
        Self {
            min_priority: Some(4),
            ..Default::default()
        }
    }
}

/// Query to get the package pool
#[derive(Debug, Clone, Default)]
pub struct Query {
    /// Filter criteria
    pub filter: PackageFilter,
    /// Maximum number of results to return
    pub limit: Option<usize>,
    /// Offset for pagination
    pub offset: Option<usize>,
}

impl Query {
    /// Create a new GetPackagePool query
    pub fn new(filter: PackageFilter) -> Self {
        Self {
            filter,
            limit: None,
            offset: None,
        }
    }

    /// Create a query with pagination
    pub fn with_pagination(filter: PackageFilter, limit: usize, offset: usize) -> Self {
        Self {
            filter,
            limit: Some(limit),
            offset: Some(offset),
        }
    }

    /// Create a query for packages in pool
    pub fn in_pool() -> Self {
        Self::new(PackageFilter::in_pool())
    }

    /// Create a query for packages by courier
    pub fn by_courier(courier_id: Uuid) -> Self {
        Self::new(PackageFilter::by_courier(courier_id))
    }
}
