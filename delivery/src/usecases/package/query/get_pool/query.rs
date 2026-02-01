//! Get Package Pool Query
//!
//! Data structure representing the query to retrieve packages.

use uuid::Uuid;

use crate::domain::model::package::PackageStatus;

/// Filter criteria for package pool
#[derive(Debug, Clone, Default)]
pub struct PackageFilter {
    /// Filter by single status
    pub status: Option<PackageStatus>,
    /// Filter by multiple statuses
    pub statuses: Option<Vec<PackageStatus>>,
    /// Filter by delivery zone
    pub zone: Option<String>,
    /// Filter by assigned courier
    pub courier_id: Option<Uuid>,
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

    /// Create a filter with multiple statuses
    pub fn with_statuses(statuses: Vec<PackageStatus>) -> Self {
        Self {
            statuses: Some(statuses),
            ..Default::default()
        }
    }

    /// Add zone filter
    pub fn and_zone(mut self, zone: &str) -> Self {
        self.zone = Some(zone.to_string());
        self
    }

    /// Add courier filter
    pub fn and_courier(mut self, courier_id: Uuid) -> Self {
        self.courier_id = Some(courier_id);
        self
    }
}

/// Query to get the package pool
#[derive(Debug, Clone, Default)]
pub struct Query {
    /// Filter criteria
    pub filter: PackageFilter,
    /// Page number (1-based)
    pub page: Option<u32>,
    /// Number of items per page
    pub page_size: Option<u32>,
}

impl Query {
    /// Create a new GetPackagePool query
    pub fn new(filter: PackageFilter) -> Self {
        Self {
            filter,
            page: None,
            page_size: None,
        }
    }

    /// Create a query with pagination
    pub fn with_pagination(filter: PackageFilter, page: u32, page_size: u32) -> Self {
        Self {
            filter,
            page: Some(page),
            page_size: Some(page_size),
        }
    }

    /// Create a query for packages in pool
    pub fn in_pool() -> Self {
        Self::new(PackageFilter::in_pool())
    }

    /// Create a query for packages in a specific zone
    pub fn in_zone(zone: &str) -> Self {
        Self::new(PackageFilter::in_zone(zone))
    }

    /// Create a query for packages by courier
    pub fn by_courier(courier_id: Uuid) -> Self {
        Self::new(PackageFilter::by_courier(courier_id))
    }

    /// Set pagination
    pub fn paginate(mut self, page: u32, page_size: u32) -> Self {
        self.page = Some(page);
        self.page_size = Some(page_size);
        self
    }
}
