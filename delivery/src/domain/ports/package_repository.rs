//! Package Repository Port
//!
//! Defines the interface for persisting and retrieving Package aggregates.
//! This port is implemented by infrastructure adapters (e.g., PostgreSQL).

use async_trait::async_trait;
use uuid::Uuid;

use crate::domain::model::package::{Package, PackageId, PackageStatus};

use super::RepositoryError;

/// Filter criteria for querying packages
#[derive(Debug, Clone, Default)]
pub struct PackageFilter {
    /// Filter by status
    pub status: Option<PackageStatus>,
    /// Filter by statuses (multiple)
    pub statuses: Option<Vec<PackageStatus>>,
    /// Filter by zone
    pub zone: Option<String>,
    /// Filter by courier ID
    pub courier_id: Option<Uuid>,
    /// Only unassigned packages
    pub unassigned_only: bool,
}

impl PackageFilter {
    /// Create a filter for packages in a specific zone
    pub fn in_zone(zone: &str) -> Self {
        Self {
            zone: Some(zone.to_string()),
            ..Default::default()
        }
    }

    /// Create a filter for packages with a specific status
    pub fn with_status(status: PackageStatus) -> Self {
        Self {
            status: Some(status),
            ..Default::default()
        }
    }

    /// Create a filter for unassigned packages in pool
    pub fn in_pool() -> Self {
        Self {
            status: Some(PackageStatus::InPool),
            unassigned_only: true,
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
}

/// Package Repository Port
///
/// Defines the contract for storing and retrieving Package aggregates.
/// Implementations handle the actual persistence mechanism (PostgreSQL, etc.).
#[async_trait]
pub trait PackageRepository: Send + Sync {
    /// Save a package (insert or update)
    ///
    /// If the package doesn't exist, it will be inserted.
    /// If it exists, it will be updated with optimistic locking.
    async fn save(&self, package: &Package) -> Result<(), RepositoryError>;

    /// Find a package by ID
    async fn find_by_id(&self, id: PackageId) -> Result<Option<Package>, RepositoryError>;

    /// Find a package by order ID
    async fn find_by_order_id(&self, order_id: Uuid) -> Result<Option<Package>, RepositoryError>;

    /// Find packages with filter
    async fn find_by_filter(
        &self,
        filter: PackageFilter,
        limit: u64,
        offset: u64,
    ) -> Result<Vec<Package>, RepositoryError>;

    /// Count packages matching filter
    async fn count_by_filter(&self, filter: PackageFilter) -> Result<u64, RepositoryError>;

    /// Find packages assigned to a courier
    async fn find_by_courier(&self, courier_id: Uuid) -> Result<Vec<Package>, RepositoryError>;

    /// Delete a package by ID
    async fn delete(&self, id: PackageId) -> Result<(), RepositoryError>;
}
