//! Get Package Pool Handler
//!
//! Handles retrieving packages with filtering and pagination.
//!
//! ## Flow
//! 1. Validate query parameters
//! 2. Convert query filter to repository filter
//! 3. Execute query on repository
//! 4. Return paginated results

use std::sync::Arc;

use chrono::{DateTime, Utc};
use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::package::{Address, DeliveryPeriod, PackageStatus, Priority};
use crate::domain::ports::{PackageFilter, PackageRepository, QueryHandler, RepositoryError};

use super::Query;

/// Maximum page size allowed
const MAX_PAGE_SIZE: usize = 100;

/// Default page size
const DEFAULT_PAGE_SIZE: usize = 20;

/// Errors that can occur during package pool retrieval
#[derive(Debug, Error)]
pub enum GetPackagePoolError {
    /// Invalid pagination parameters
    #[error("Invalid pagination: {0}")]
    InvalidPagination(String),

    /// Invalid date range
    #[error("Invalid date range: start must be before end")]
    InvalidDateRange,

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),
}

/// Package data from the pool
#[derive(Debug, Clone)]
pub struct PackageInfo {
    /// Package ID
    pub id: Uuid,
    /// Order ID from OMS
    pub order_id: Uuid,
    /// Customer ID
    pub customer_id: Uuid,
    /// Pickup address
    pub pickup_address: Address,
    /// Delivery address
    pub delivery_address: Address,
    /// Delivery period
    pub delivery_period: DeliveryPeriod,
    /// Weight in kg
    pub weight_kg: f64,
    /// Priority level
    pub priority: Priority,
    /// Current status
    pub status: PackageStatus,
    /// Delivery zone
    pub zone: String,
    /// Assigned courier ID (if any)
    pub courier_id: Option<Uuid>,
    /// When the package was created
    pub created_at: DateTime<Utc>,
    /// When the package was assigned
    pub assigned_at: Option<DateTime<Utc>>,
    /// When the package was delivered
    pub delivered_at: Option<DateTime<Utc>>,
    /// Reason for not delivered (if applicable)
    pub not_delivered_reason: Option<String>,
}

/// Pagination info for response
#[derive(Debug, Clone)]
pub struct PaginationInfo {
    /// Current page number (1-based)
    pub current_page: u32,
    /// Items per page
    pub page_size: u32,
    /// Total number of pages
    pub total_pages: u32,
    /// Total number of items
    pub total_items: u64,
}

/// Response from get package pool query
#[derive(Debug)]
pub struct Response {
    /// List of packages
    pub packages: Vec<PackageInfo>,
    /// Total count of matching packages (before pagination)
    pub total_count: u64,
    /// Pagination info
    pub pagination: PaginationInfo,
}

/// Get Package Pool Handler
pub struct Handler<P>
where
    P: PackageRepository,
{
    package_repo: Arc<P>,
}

impl<P> Handler<P>
where
    P: PackageRepository,
{
    /// Create a new handler instance
    pub fn new(package_repo: Arc<P>) -> Self {
        Self { package_repo }
    }

    /// Convert query filter to repository filter
    fn to_repo_filter(query: &Query) -> PackageFilter {
        let mut filter = PackageFilter::default();

        // Single status filter
        if let Some(status) = &query.filter.status {
            filter.status = Some(*status);
        }

        // Multiple statuses filter
        if let Some(statuses) = &query.filter.statuses {
            filter.statuses = Some(statuses.clone());
        }

        // Zone filter
        if let Some(zone) = &query.filter.zone {
            filter.zone = Some(zone.clone());
        }

        // Courier filter
        if let Some(courier_id) = query.filter.courier_id {
            filter.courier_id = Some(courier_id);
        }

        // Unassigned only
        filter.unassigned_only = query.filter.unassigned_only;

        filter
    }

    /// Calculate pagination values
    fn calculate_pagination(
        page: Option<u32>,
        page_size: Option<u32>,
        total_count: u64,
    ) -> Result<(u64, u64, PaginationInfo), GetPackagePoolError> {
        let page = page.unwrap_or(1);
        let page_size = page_size.unwrap_or(DEFAULT_PAGE_SIZE as u32);

        if page == 0 {
            return Err(GetPackagePoolError::InvalidPagination(
                "Page must be >= 1".to_string(),
            ));
        }

        if page_size == 0 {
            return Err(GetPackagePoolError::InvalidPagination(
                "Page size must be >= 1".to_string(),
            ));
        }

        let page_size = page_size.min(MAX_PAGE_SIZE as u32);
        let offset = (page - 1) as u64 * page_size as u64;
        let limit = page_size as u64;

        let total_pages = if total_count == 0 {
            1
        } else {
            ((total_count as f64) / (page_size as f64)).ceil() as u32
        };

        let pagination = PaginationInfo {
            current_page: page,
            page_size,
            total_pages,
            total_items: total_count,
        };

        Ok((limit, offset, pagination))
    }
}

impl<P> QueryHandler<Query, Response> for Handler<P>
where
    P: PackageRepository + Send + Sync,
{
    type Error = GetPackagePoolError;

    /// Handle the GetPackagePool query
    async fn handle(&self, query: Query) -> Result<Response, Self::Error> {
        // 1. Convert query filter to repository filter
        let repo_filter = Self::to_repo_filter(&query);

        // 2. Get total count for pagination
        let total_count = self.package_repo.count_by_filter(repo_filter.clone()).await?;

        // 3. Calculate pagination
        let (limit, offset, pagination) =
            Self::calculate_pagination(query.page, query.page_size, total_count)?;

        // 4. Execute query with pagination
        let packages = self
            .package_repo
            .find_by_filter(repo_filter, limit, offset)
            .await?;

        // 5. Map to response
        let packages: Vec<PackageInfo> = packages
            .into_iter()
            .map(|p| PackageInfo {
                id: p.id().0,
                order_id: p.order_id(),
                customer_id: p.customer_id(),
                pickup_address: p.pickup_address().clone(),
                delivery_address: p.delivery_address().clone(),
                delivery_period: p.delivery_period().clone(),
                weight_kg: p.weight_kg(),
                priority: p.priority(),
                status: p.status(),
                zone: p.zone().to_string(),
                courier_id: p.courier_id(),
                created_at: p.created_at(),
                assigned_at: p.assigned_at(),
                delivered_at: p.delivered_at(),
                not_delivered_reason: p.not_delivered_reason().map(|s| s.to_string()),
            })
            .collect();

        Ok(Response {
            packages,
            total_count,
            pagination,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::package::{Package, PackageId};
    use crate::domain::model::vo::location::Location;
    use async_trait::async_trait;
    use std::collections::HashMap;
    use std::sync::Mutex;

    // ==================== Mock Repository ====================

    struct MockPackageRepository {
        packages: Mutex<HashMap<Uuid, Package>>,
    }

    impl MockPackageRepository {
        fn new() -> Self {
            Self {
                packages: Mutex::new(HashMap::new()),
            }
        }

        fn add_package(&self, package: Package) {
            let mut packages = self.packages.lock().unwrap();
            packages.insert(package.id().0, package);
        }
    }

    #[async_trait]
    impl PackageRepository for MockPackageRepository {
        async fn save(&self, package: &Package) -> Result<(), RepositoryError> {
            let mut packages = self.packages.lock().unwrap();
            packages.insert(package.id().0, package.clone());
            Ok(())
        }

        async fn find_by_id(&self, id: PackageId) -> Result<Option<Package>, RepositoryError> {
            let packages = self.packages.lock().unwrap();
            Ok(packages.get(&id.0).cloned())
        }

        async fn find_by_order_id(&self, _order_id: Uuid) -> Result<Option<Package>, RepositoryError> {
            Ok(None)
        }

        async fn find_by_filter(
            &self,
            filter: PackageFilter,
            limit: u64,
            offset: u64,
        ) -> Result<Vec<Package>, RepositoryError> {
            let packages = self.packages.lock().unwrap();
            let mut result: Vec<Package> = packages
                .values()
                .filter(|p| {
                    // Apply status filter
                    if let Some(status) = &filter.status {
                        if p.status() != *status {
                            return false;
                        }
                    }

                    // Apply statuses filter
                    if let Some(statuses) = &filter.statuses {
                        if !statuses.contains(&p.status()) {
                            return false;
                        }
                    }

                    // Apply zone filter
                    if let Some(zone) = &filter.zone {
                        if p.zone() != zone {
                            return false;
                        }
                    }

                    // Apply courier filter
                    if let Some(courier_id) = filter.courier_id {
                        if p.courier_id() != Some(courier_id) {
                            return false;
                        }
                    }

                    // Apply unassigned only
                    if filter.unassigned_only && p.courier_id().is_some() {
                        return false;
                    }

                    true
                })
                .cloned()
                .collect();

            // Sort by created_at descending
            result.sort_by(|a, b| b.created_at().cmp(&a.created_at()));

            // Apply pagination
            let offset = offset as usize;
            let limit = limit as usize;
            let result = result.into_iter().skip(offset).take(limit).collect();

            Ok(result)
        }

        async fn count_by_filter(&self, filter: PackageFilter) -> Result<u64, RepositoryError> {
            let packages = self.packages.lock().unwrap();
            let count = packages
                .values()
                .filter(|p| {
                    // Apply same filters as find_by_filter
                    if let Some(status) = &filter.status {
                        if p.status() != *status {
                            return false;
                        }
                    }

                    if let Some(statuses) = &filter.statuses {
                        if !statuses.contains(&p.status()) {
                            return false;
                        }
                    }

                    if let Some(zone) = &filter.zone {
                        if p.zone() != zone {
                            return false;
                        }
                    }

                    if let Some(courier_id) = filter.courier_id {
                        if p.courier_id() != Some(courier_id) {
                            return false;
                        }
                    }

                    if filter.unassigned_only && p.courier_id().is_some() {
                        return false;
                    }

                    true
                })
                .count();

            Ok(count as u64)
        }

        async fn find_by_courier(&self, courier_id: Uuid) -> Result<Vec<Package>, RepositoryError> {
            let packages = self.packages.lock().unwrap();
            Ok(packages
                .values()
                .filter(|p| p.courier_id() == Some(courier_id))
                .cloned()
                .collect())
        }

        async fn delete(&self, _id: PackageId) -> Result<(), RepositoryError> {
            Ok(())
        }
    }

    // ==================== Test Helpers ====================

    fn create_test_address() -> Address {
        Address::new(
            "123 Main St".to_string(),
            "Berlin".to_string(),
            "10115".to_string(),
            Location::new(52.52, 13.405, 10.0).unwrap(),
        )
    }

    fn create_test_package_in_pool(zone: &str) -> Package {
        let now = Utc::now();
        let period = DeliveryPeriod::new(
            now + chrono::Duration::hours(2),
            now + chrono::Duration::hours(4),
        )
        .unwrap();

        let mut package = Package::new(
            Uuid::new_v4(),
            Uuid::new_v4(),
            None, // customer_phone
            None,
            None,
            None,
            create_test_address(),
            create_test_address(),
            period,
            2.5,
            Priority::Normal,
            zone.to_string(),
        );

        package.move_to_pool().unwrap();
        package
    }

    fn create_assigned_package(zone: &str, courier_id: Uuid) -> Package {
        let now = Utc::now();
        let period = DeliveryPeriod::new(
            now + chrono::Duration::hours(2),
            now + chrono::Duration::hours(4),
        )
        .unwrap();

        let mut package = Package::new(
            Uuid::new_v4(),
            Uuid::new_v4(),
            None, // customer_phone
            None,
            None,
            None,
            create_test_address(),
            create_test_address(),
            period,
            2.5,
            Priority::Normal,
            zone.to_string(),
        );

        package.move_to_pool().unwrap();
        package.assign_to(courier_id).unwrap();
        package
    }

    // ==================== Tests ====================

    #[tokio::test]
    async fn test_get_pool_empty() {
        let package_repo = Arc::new(MockPackageRepository::new());
        let handler = Handler::new(package_repo);

        let query = Query::in_pool();
        let result = handler.handle(query).await;

        assert!(result.is_ok());
        let response = result.unwrap();
        assert_eq!(response.packages.len(), 0);
        assert_eq!(response.total_count, 0);
        assert_eq!(response.pagination.current_page, 1);
    }

    #[tokio::test]
    async fn test_get_pool_returns_packages() {
        let package_repo = Arc::new(MockPackageRepository::new());

        // Add some packages
        for _ in 0..5 {
            package_repo.add_package(create_test_package_in_pool("Berlin-101"));
        }

        let handler = Handler::new(package_repo);

        let query = Query::in_pool();
        let result = handler.handle(query).await;

        assert!(result.is_ok());
        let response = result.unwrap();
        assert_eq!(response.packages.len(), 5);
        assert_eq!(response.total_count, 5);
    }

    #[tokio::test]
    async fn test_get_pool_filter_by_zone() {
        let package_repo = Arc::new(MockPackageRepository::new());

        // Add packages in different zones
        for _ in 0..3 {
            package_repo.add_package(create_test_package_in_pool("Berlin-101"));
        }
        for _ in 0..2 {
            package_repo.add_package(create_test_package_in_pool("Munich-201"));
        }

        let handler = Handler::new(package_repo);

        let query = Query::in_zone("Berlin-101");
        let result = handler.handle(query).await;

        assert!(result.is_ok());
        let response = result.unwrap();
        assert_eq!(response.packages.len(), 3);
        assert_eq!(response.total_count, 3);
    }

    #[tokio::test]
    async fn test_get_pool_filter_by_courier() {
        let package_repo = Arc::new(MockPackageRepository::new());

        let courier_id = Uuid::new_v4();
        let other_courier_id = Uuid::new_v4();

        // Add packages assigned to different couriers
        for _ in 0..3 {
            package_repo.add_package(create_assigned_package("Berlin-101", courier_id));
        }
        for _ in 0..2 {
            package_repo.add_package(create_assigned_package("Berlin-101", other_courier_id));
        }

        let handler = Handler::new(package_repo);

        let query = Query::by_courier(courier_id);
        let result = handler.handle(query).await;

        assert!(result.is_ok());
        let response = result.unwrap();
        assert_eq!(response.packages.len(), 3);
        assert_eq!(response.total_count, 3);
    }

    #[tokio::test]
    async fn test_get_pool_pagination() {
        let package_repo = Arc::new(MockPackageRepository::new());

        // Add 25 packages
        for _ in 0..25 {
            package_repo.add_package(create_test_package_in_pool("Berlin-101"));
        }

        let handler = Handler::new(package_repo);

        // First page
        let query = Query {
            filter: super::super::query::PackageFilter::in_pool(),
            page: Some(1),
            page_size: Some(10),
            ..Default::default()
        };
        let result = handler.handle(query).await;

        assert!(result.is_ok());
        let response = result.unwrap();
        assert_eq!(response.packages.len(), 10);
        assert_eq!(response.total_count, 25);
        assert_eq!(response.pagination.current_page, 1);
        assert_eq!(response.pagination.page_size, 10);
        assert_eq!(response.pagination.total_pages, 3);
        assert_eq!(response.pagination.total_items, 25);

        // Second page
        let query = Query {
            filter: super::super::query::PackageFilter::in_pool(),
            page: Some(2),
            page_size: Some(10),
            ..Default::default()
        };
        let result = handler.handle(query).await;

        assert!(result.is_ok());
        let response = result.unwrap();
        assert_eq!(response.packages.len(), 10);
        assert_eq!(response.pagination.current_page, 2);

        // Third page (partial)
        let query = Query {
            filter: super::super::query::PackageFilter::in_pool(),
            page: Some(3),
            page_size: Some(10),
            ..Default::default()
        };
        let result = handler.handle(query).await;

        assert!(result.is_ok());
        let response = result.unwrap();
        assert_eq!(response.packages.len(), 5);
        assert_eq!(response.pagination.current_page, 3);
    }

    #[tokio::test]
    async fn test_get_pool_invalid_pagination() {
        let package_repo = Arc::new(MockPackageRepository::new());
        let handler = Handler::new(package_repo);

        // Page 0 is invalid
        let query = Query {
            filter: super::super::query::PackageFilter::default(),
            page: Some(0),
            page_size: Some(10),
            ..Default::default()
        };
        let result = handler.handle(query).await;

        assert!(matches!(result, Err(GetPackagePoolError::InvalidPagination(_))));
    }

    #[tokio::test]
    async fn test_get_pool_max_page_size() {
        let package_repo = Arc::new(MockPackageRepository::new());

        // Add 150 packages
        for _ in 0..150 {
            package_repo.add_package(create_test_package_in_pool("Berlin-101"));
        }

        let handler = Handler::new(package_repo);

        // Request page size > MAX_PAGE_SIZE
        let query = Query {
            filter: super::super::query::PackageFilter::in_pool(),
            page: Some(1),
            page_size: Some(500), // More than MAX_PAGE_SIZE
            ..Default::default()
        };
        let result = handler.handle(query).await;

        assert!(result.is_ok());
        let response = result.unwrap();
        // Should be capped at MAX_PAGE_SIZE (100)
        assert_eq!(response.packages.len(), 100);
        assert_eq!(response.pagination.page_size, 100);
    }
}
