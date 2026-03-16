//! Delivery Activities
//!
//! Temporal activities for delivery operations.
//!
//! These activities are registered with the Temporal worker in `runner.rs`
//! and called from delivery workflows (assign_order, deliver_order).

use std::sync::Arc;

use thiserror::Error;
use tracing::{info, warn};
use uuid::Uuid;

use crate::domain::model::package::PackageStatus;
use crate::domain::ports::{
    CommandHandlerWithResult, CourierCache, CourierRepository, PackageRepository, QueryHandler,
};
use crate::domain::services::dispatch::DispatchCandidate;
use crate::usecases::courier::command::{
    AcceptPackageCommand, AcceptPackageError, AcceptPackageHandler, CompleteCourierDeliveryCommand,
    CompleteCourierDeliveryError, CompleteCourierDeliveryHandler,
};
use crate::usecases::courier::query::get_pool::{
    GetCourierPoolError, Handler as GetPoolHandler, Query as GetPoolQuery,
};

/// Errors from delivery activities
#[derive(Debug, Error)]
pub enum DeliveryActivityError {
    #[error("Repository error: {0}")]
    RepositoryError(String),

    #[error("Cache error: {0}")]
    CacheError(String),

    #[error("No couriers available in zone: {0}")]
    NoCouriersAvailable(String),

    #[error("Courier became unavailable during activity: {0}")]
    CourierUnavailable(Uuid),

    #[error("Dispatch failed: {0}")]
    DispatchFailed(String),
}

/// Delivery Activities - thin wrappers around use cases
pub struct DeliveryActivities<R, C, P>
where
    R: CourierRepository,
    C: CourierCache,
    P: PackageRepository,
{
    get_pool_handler: Arc<GetPoolHandler<R, C>>,
    accept_package_handler: Arc<AcceptPackageHandler<R, C>>,
    complete_delivery_handler: Arc<CompleteCourierDeliveryHandler<R, C>>,
    package_repo: Arc<P>,
}

impl<R, C, P> DeliveryActivities<R, C, P>
where
    R: CourierRepository + Send + Sync,
    C: CourierCache + Send + Sync,
    P: PackageRepository + Send + Sync,
{
    /// Create new delivery activities
    pub fn new(
        get_pool_handler: Arc<GetPoolHandler<R, C>>,
        accept_package_handler: Arc<AcceptPackageHandler<R, C>>,
        complete_delivery_handler: Arc<CompleteCourierDeliveryHandler<R, C>>,
        package_repo: Arc<P>,
    ) -> Self {
        Self {
            get_pool_handler,
            accept_package_handler,
            complete_delivery_handler,
            package_repo,
        }
    }

    // =========================================================================
    // Activity: Get Dispatch Candidates
    // =========================================================================

    /// Get dispatch candidates in a zone.
    ///
    /// Note: candidates are returned without current location. Location fetching
    /// is handled separately before nearest-courier selection.
    pub async fn get_dispatch_candidates(
        &self,
        zone: &str,
    ) -> Result<Vec<DispatchCandidate>, DeliveryActivityError> {
        let query = GetPoolQuery::free_in_zone(zone);

        let result = self
            .get_pool_handler
            .handle(query)
            .await
            .map_err(Self::map_get_pool_error)?;

        let couriers: Vec<DispatchCandidate> = result
            .couriers
            .into_iter()
            .map(|cws| DispatchCandidate {
                courier: cws.courier,
                current_location: None,
            })
            .collect();

        if couriers.is_empty() {
            return Err(DeliveryActivityError::NoCouriersAvailable(zone.to_string()));
        }

        Ok(couriers)
    }

    // =========================================================================
    // Activity: Assign Order to Courier
    // =========================================================================

    /// Assign an order to a courier
    ///
    /// Updates courier state in cache.
    pub async fn assign_order(
        &self,
        courier_id: Uuid,
        order_id: Uuid,
    ) -> Result<(), DeliveryActivityError> {
        info!(
            courier_id = %courier_id,
            order_id = %order_id,
            phase = "started",
            "Starting assign_order activity"
        );

        if self
            .is_assign_order_already_applied(courier_id, order_id)
            .await?
        {
            info!(
                courier_id = %courier_id,
                order_id = %order_id,
                phase = "already_applied",
                "assign_order activity already applied, treating retry as success"
            );
            return Ok(());
        }

        let result = self
            .accept_package_handler
            .handle(AcceptPackageCommand::new(courier_id))
            .await
            .map(|_| ())
            .map_err(Self::map_accept_package_error);

        match &result {
            Ok(()) => info!(
                courier_id = %courier_id,
                order_id = %order_id,
                phase = "completed",
                "assign_order activity completed"
            ),
            Err(error) => warn!(
                courier_id = %courier_id,
                order_id = %order_id,
                phase = "failed",
                error = %error,
                "assign_order activity failed"
            ),
        }

        result
    }

    // =========================================================================
    // Activity: Complete Delivery
    // =========================================================================

    /// Complete a delivery (success or failure)
    pub async fn complete_delivery(
        &self,
        courier_id: Uuid,
        order_id: Uuid,
        success: bool,
    ) -> Result<(), DeliveryActivityError> {
        info!(
            courier_id = %courier_id,
            order_id = %order_id,
            success,
            phase = "started",
            "Starting complete_delivery activity"
        );

        if self
            .is_complete_delivery_already_applied(courier_id, order_id, success)
            .await?
        {
            info!(
                courier_id = %courier_id,
                order_id = %order_id,
                success,
                phase = "already_applied",
                "complete_delivery activity already applied, treating retry as success"
            );
            return Ok(());
        }

        let result = self
            .complete_delivery_handler
            .handle(CompleteCourierDeliveryCommand::new(courier_id, success))
            .await
            .map(|_| ())
            .map_err(Self::map_complete_delivery_error);

        match &result {
            Ok(()) => info!(
                courier_id = %courier_id,
                order_id = %order_id,
                success,
                phase = "completed",
                "complete_delivery activity completed"
            ),
            Err(error) => warn!(
                courier_id = %courier_id,
                order_id = %order_id,
                success,
                phase = "failed",
                error = %error,
                "complete_delivery activity failed"
            ),
        }

        result
    }

    fn map_accept_package_error(error: AcceptPackageError) -> DeliveryActivityError {
        match error {
            AcceptPackageError::NotFound(courier_id) => {
                DeliveryActivityError::CourierUnavailable(courier_id)
            }
            AcceptPackageError::DomainError(error) => {
                DeliveryActivityError::DispatchFailed(error.to_string())
            }
            AcceptPackageError::RepositoryError(error) => {
                DeliveryActivityError::RepositoryError(error.to_string())
            }
        }
    }

    fn map_complete_delivery_error(error: CompleteCourierDeliveryError) -> DeliveryActivityError {
        match error {
            CompleteCourierDeliveryError::NotFound(courier_id) => {
                DeliveryActivityError::CourierUnavailable(courier_id)
            }
            CompleteCourierDeliveryError::DomainError(error) => {
                DeliveryActivityError::DispatchFailed(error.to_string())
            }
            CompleteCourierDeliveryError::RepositoryError(error) => {
                DeliveryActivityError::RepositoryError(error.to_string())
            }
        }
    }

    fn map_get_pool_error(error: GetCourierPoolError) -> DeliveryActivityError {
        match error {
            GetCourierPoolError::RepositoryError(error) => {
                DeliveryActivityError::RepositoryError(error.to_string())
            }
            GetCourierPoolError::CacheError(error) => {
                DeliveryActivityError::CacheError(error.to_string())
            }
        }
    }

    async fn is_assign_order_already_applied(
        &self,
        courier_id: Uuid,
        order_id: Uuid,
    ) -> Result<bool, DeliveryActivityError> {
        let Some(package) = self
            .package_repo
            .find_by_order_id(order_id)
            .await
            .map_err(|error| DeliveryActivityError::RepositoryError(error.to_string()))?
        else {
            return Ok(false);
        };

        Ok(package.courier_id() == Some(courier_id)
            && matches!(
                package.status(),
                PackageStatus::Assigned
                    | PackageStatus::InTransit
                    | PackageStatus::Delivered
                    | PackageStatus::NotDelivered
            ))
    }

    async fn is_complete_delivery_already_applied(
        &self,
        courier_id: Uuid,
        order_id: Uuid,
        success: bool,
    ) -> Result<bool, DeliveryActivityError> {
        let Some(package) = self
            .package_repo
            .find_by_order_id(order_id)
            .await
            .map_err(|error| DeliveryActivityError::RepositoryError(error.to_string()))?
        else {
            return Ok(false);
        };

        let expected_status = if success {
            PackageStatus::Delivered
        } else {
            PackageStatus::NotDelivered
        };

        Ok(package.courier_id() == Some(courier_id) && package.status() == expected_status)
    }
}

// Note: Activity registration is done in `runner.rs` using the Temporal SDK.
// The methods here use domain/application types; `runner.rs` provides the
// Temporal-facing wrappers that serialize activity inputs/outputs as strings
// for compatibility with the current pre-alpha SDK.

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::courier::{Courier, CourierId, CourierStatus, WorkHours};
    use crate::domain::model::package::{
        Address, DeliveryPeriod, NotDeliveredDetails, NotDeliveredReasonCode, Package, PackageId,
        Priority,
    };
    use crate::domain::model::vo::location::Location;
    use crate::domain::model::vo::TransportType;
    use crate::domain::ports::{
        CacheError, CachedCourierState, PackageFilter, PackageRepository, RepositoryError,
    };
    use async_trait::async_trait;
    use chrono::{NaiveTime, Utc};
    use std::collections::HashMap;
    use std::sync::Mutex;

    struct MockCourierRepository {
        couriers: Mutex<HashMap<Uuid, Courier>>,
        fail_save: bool,
        save_calls: Mutex<u32>,
    }

    impl MockCourierRepository {
        fn new() -> Self {
            Self {
                couriers: Mutex::new(HashMap::new()),
                fail_save: false,
                save_calls: Mutex::new(0),
            }
        }

        fn failing_on_save() -> Self {
            Self {
                fail_save: true,
                ..Self::new()
            }
        }

        fn add_courier(&self, courier: Courier) {
            self.couriers
                .lock()
                .unwrap()
                .insert(courier.id().0, courier);
        }

        fn saved_courier(&self, courier_id: Uuid) -> Option<Courier> {
            self.couriers.lock().unwrap().get(&courier_id).cloned()
        }

        fn save_calls(&self) -> u32 {
            *self.save_calls.lock().unwrap()
        }
    }

    #[async_trait]
    impl CourierRepository for MockCourierRepository {
        async fn save(&self, courier: &Courier) -> Result<(), RepositoryError> {
            *self.save_calls.lock().unwrap() += 1;
            if self.fail_save {
                return Err(RepositoryError::QueryError(
                    "simulated repository save failure".to_string(),
                ));
            }

            self.couriers
                .lock()
                .unwrap()
                .insert(courier.id().0, courier.clone());
            Ok(())
        }

        async fn find_by_id(&self, id: Uuid) -> Result<Option<Courier>, RepositoryError> {
            Ok(self.couriers.lock().unwrap().get(&id).cloned())
        }

        async fn find_by_phone(&self, _phone: &str) -> Result<Option<Courier>, RepositoryError> {
            Ok(None)
        }

        async fn find_by_email(&self, _email: &str) -> Result<Option<Courier>, RepositoryError> {
            Ok(None)
        }

        async fn find_by_work_zone(&self, _zone: &str) -> Result<Vec<Courier>, RepositoryError> {
            Ok(vec![])
        }

        async fn email_exists(&self, _email: &str) -> Result<bool, RepositoryError> {
            Ok(false)
        }

        async fn phone_exists(&self, _phone: &str) -> Result<bool, RepositoryError> {
            Ok(false)
        }

        async fn delete(&self, _id: Uuid) -> Result<(), RepositoryError> {
            Ok(())
        }

        async fn archive(&self, _id: Uuid) -> Result<(), RepositoryError> {
            Ok(())
        }

        async fn list(&self, _limit: u64, _offset: u64) -> Result<Vec<Courier>, RepositoryError> {
            Ok(vec![])
        }

        async fn find_by_filter(
            &self,
            _filter: crate::domain::ports::CourierFilter,
            _limit: u64,
            _offset: u64,
        ) -> Result<Vec<Courier>, RepositoryError> {
            Ok(self.couriers.lock().unwrap().values().cloned().collect())
        }

        async fn count_by_filter(
            &self,
            _filter: crate::domain::ports::CourierFilter,
        ) -> Result<u64, RepositoryError> {
            Ok(self.couriers.lock().unwrap().len() as u64)
        }
    }

    struct MockCourierCache {
        fail_cache: bool,
        cache_calls: Mutex<u32>,
        cached: Mutex<HashMap<Uuid, Courier>>,
    }

    impl MockCourierCache {
        fn new() -> Self {
            Self {
                fail_cache: false,
                cache_calls: Mutex::new(0),
                cached: Mutex::new(HashMap::new()),
            }
        }

        fn failing_on_cache() -> Self {
            Self {
                fail_cache: true,
                ..Self::new()
            }
        }

        fn cache_calls(&self) -> u32 {
            *self.cache_calls.lock().unwrap()
        }

        fn cached_courier(&self, courier_id: Uuid) -> Option<Courier> {
            self.cached.lock().unwrap().get(&courier_id).cloned()
        }
    }

    #[async_trait]
    impl CourierCache for MockCourierCache {
        async fn cache(&self, courier: &Courier) -> Result<(), CacheError> {
            *self.cache_calls.lock().unwrap() += 1;
            if self.fail_cache {
                return Err(CacheError::OperationError(
                    "simulated cache failure".to_string(),
                ));
            }

            self.cached
                .lock()
                .unwrap()
                .insert(courier.id().0, courier.clone());
            Ok(())
        }

        async fn get_state(
            &self,
            _courier_id: Uuid,
        ) -> Result<Option<CachedCourierState>, CacheError> {
            Ok(None)
        }

        async fn get_status(&self, _courier_id: Uuid) -> Result<Option<CourierStatus>, CacheError> {
            Ok(None)
        }

        async fn get_free_couriers_in_zone(&self, _zone: &str) -> Result<Vec<Uuid>, CacheError> {
            Ok(vec![])
        }

        async fn get_all_free_couriers(&self) -> Result<Vec<Uuid>, CacheError> {
            Ok(vec![])
        }

        async fn remove(&self, _courier_id: Uuid, _work_zone: &str) -> Result<(), CacheError> {
            Ok(())
        }

        async fn exists(&self, _courier_id: Uuid) -> Result<bool, CacheError> {
            Ok(true)
        }
    }

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
            self.packages
                .lock()
                .unwrap()
                .insert(package.order_id(), package);
        }
    }

    #[async_trait]
    impl PackageRepository for MockPackageRepository {
        async fn save(&self, package: &Package) -> Result<(), RepositoryError> {
            self.packages
                .lock()
                .unwrap()
                .insert(package.order_id(), package.clone());
            Ok(())
        }

        async fn find_by_id(&self, id: PackageId) -> Result<Option<Package>, RepositoryError> {
            Ok(self
                .packages
                .lock()
                .unwrap()
                .values()
                .find(|package| package.id().0 == id.0)
                .cloned())
        }

        async fn find_by_order_id(
            &self,
            order_id: Uuid,
        ) -> Result<Option<Package>, RepositoryError> {
            Ok(self.packages.lock().unwrap().get(&order_id).cloned())
        }

        async fn find_by_filter(
            &self,
            _filter: PackageFilter,
            _limit: u64,
            _offset: u64,
        ) -> Result<Vec<Package>, RepositoryError> {
            Ok(vec![])
        }

        async fn count_by_filter(&self, _filter: PackageFilter) -> Result<u64, RepositoryError> {
            Ok(0)
        }

        async fn find_by_courier(
            &self,
            _courier_id: Uuid,
        ) -> Result<Vec<Package>, RepositoryError> {
            Ok(vec![])
        }

        async fn delete(&self, _id: PackageId) -> Result<(), RepositoryError> {
            Ok(())
        }
    }

    fn create_activities(
        courier_repo: Arc<MockCourierRepository>,
        courier_cache: Arc<MockCourierCache>,
        package_repo: Arc<MockPackageRepository>,
    ) -> DeliveryActivities<MockCourierRepository, MockCourierCache, MockPackageRepository> {
        let get_pool_handler = Arc::new(GetPoolHandler::new(
            courier_repo.clone(),
            courier_cache.clone(),
        ));
        let accept_package_handler = Arc::new(AcceptPackageHandler::new(
            courier_repo.clone(),
            courier_cache.clone(),
        ));
        let complete_delivery_handler = Arc::new(CompleteCourierDeliveryHandler::new(
            courier_repo.clone(),
            courier_cache.clone(),
        ));
        DeliveryActivities::new(
            get_pool_handler,
            accept_package_handler,
            complete_delivery_handler,
            package_repo,
        )
    }

    fn create_test_address() -> Address {
        Address::new(
            "123 Main St".to_string(),
            "Berlin".to_string(),
            Location::new(52.52, 13.405, 10.0).unwrap(),
        )
    }

    fn create_test_courier(id: Uuid, status: CourierStatus, current_load: u32) -> Courier {
        let work_hours = WorkHours::new(
            NaiveTime::from_hms_opt(9, 0, 0).unwrap(),
            NaiveTime::from_hms_opt(18, 0, 0).unwrap(),
            vec![1, 2, 3, 4, 5],
        )
        .unwrap();

        Courier::reconstitute(
            CourierId::from_uuid(id),
            "Test Courier".to_string(),
            "+1234567890".to_string(),
            "test@example.com".to_string(),
            TransportType::Car,
            50.0,
            "Berlin-101".to_string(),
            work_hours,
            None,
            status,
            current_load,
            4.5,
            10,
            1,
            Utc::now(),
            Utc::now(),
            1,
        )
        .unwrap()
    }

    fn create_test_package(
        order_id: Uuid,
        courier_id: Option<Uuid>,
        status: PackageStatus,
    ) -> Package {
        let now = Utc::now();
        let period = DeliveryPeriod::new(
            now + chrono::Duration::hours(2),
            now + chrono::Duration::hours(4),
        )
        .unwrap();

        let mut package = Package::new(
            order_id,
            Uuid::new_v4(),
            None,
            None,
            None,
            None,
            create_test_address(),
            create_test_address(),
            period,
            2.5,
            Priority::Normal,
            "Berlin-101".to_string(),
        );

        package.move_to_pool().unwrap();

        if let Some(courier_id) = courier_id {
            package.assign_to(courier_id).unwrap();
        }

        match status {
            PackageStatus::Assigned => {}
            PackageStatus::InTransit => package.start_transit().unwrap(),
            PackageStatus::Delivered => {
                package.start_transit().unwrap();
                package.mark_delivered().unwrap();
            }
            PackageStatus::NotDelivered => {
                package.start_transit().unwrap();
                package
                    .mark_not_delivered(NotDeliveredDetails::new(
                        NotDeliveredReasonCode::CustomerUnavailable,
                        None,
                    ))
                    .unwrap();
            }
            other => panic!("unsupported test package status: {:?}", other),
        }

        package
    }

    #[tokio::test]
    async fn get_dispatch_candidates_returns_no_couriers_available_on_empty_pool() {
        let repository = Arc::new(MockCourierRepository::new());
        let cache = Arc::new(MockCourierCache::new());
        let package_repository = Arc::new(MockPackageRepository::new());

        let activities = create_activities(repository, cache, package_repository);
        let result = activities.get_dispatch_candidates("Berlin-101").await;

        assert!(matches!(
            result,
            Err(DeliveryActivityError::NoCouriersAvailable(zone)) if zone == "Berlin-101"
        ));
    }

    #[tokio::test]
    async fn assign_order_cache_failure_is_non_fatal_after_persistence() {
        let repository = Arc::new(MockCourierRepository::new());
        let cache = Arc::new(MockCourierCache::failing_on_cache());
        let package_repository = Arc::new(MockPackageRepository::new());
        let courier_id = Uuid::new_v4();
        let order_id = Uuid::new_v4();
        repository.add_courier(create_test_courier(courier_id, CourierStatus::Free, 0));

        let activities = create_activities(repository.clone(), cache.clone(), package_repository);
        let result = activities.assign_order(courier_id, order_id).await;

        assert!(result.is_ok(), "Expected Ok, got: {:?}", result.err());
        assert_eq!(repository.save_calls(), 1);
        assert_eq!(cache.cache_calls(), 1);
        assert!(cache.cached_courier(courier_id).is_none());

        let persisted = repository.saved_courier(courier_id).unwrap();
        assert_eq!(persisted.current_load(), 1);
    }

    #[tokio::test]
    async fn complete_delivery_cache_failure_is_non_fatal_after_persistence() {
        let repository = Arc::new(MockCourierRepository::new());
        let cache = Arc::new(MockCourierCache::failing_on_cache());
        let package_repository = Arc::new(MockPackageRepository::new());
        let courier_id = Uuid::new_v4();
        let order_id = Uuid::new_v4();
        repository.add_courier(create_test_courier(courier_id, CourierStatus::Busy, 1));

        let activities = create_activities(repository.clone(), cache.clone(), package_repository);
        let result = activities
            .complete_delivery(courier_id, order_id, true)
            .await;

        assert!(result.is_ok(), "Expected Ok, got: {:?}", result.err());
        assert_eq!(repository.save_calls(), 1);
        assert_eq!(cache.cache_calls(), 1);
        assert!(cache.cached_courier(courier_id).is_none());

        let persisted = repository.saved_courier(courier_id).unwrap();
        assert_eq!(persisted.current_load(), 0);
        assert_eq!(persisted.successful_deliveries(), 11);
        assert_eq!(persisted.status(), CourierStatus::Free);
    }

    #[tokio::test]
    async fn assign_order_repository_failure_still_returns_error_and_skips_cache() {
        let repository = Arc::new(MockCourierRepository::failing_on_save());
        let cache = Arc::new(MockCourierCache::new());
        let package_repository = Arc::new(MockPackageRepository::new());
        let courier_id = Uuid::new_v4();
        let order_id = Uuid::new_v4();
        repository.add_courier(create_test_courier(courier_id, CourierStatus::Free, 0));

        let activities = create_activities(repository.clone(), cache.clone(), package_repository);
        let result = activities.assign_order(courier_id, order_id).await;

        assert!(matches!(
            result,
            Err(DeliveryActivityError::RepositoryError(_))
        ));
        assert_eq!(repository.save_calls(), 1);
        assert_eq!(cache.cache_calls(), 0);

        let persisted = repository.saved_courier(courier_id).unwrap();
        assert_eq!(persisted.current_load(), 0);
    }

    #[tokio::test]
    async fn assign_order_returns_ok_when_package_is_already_assigned_to_same_courier() {
        let repository = Arc::new(MockCourierRepository::new());
        let cache = Arc::new(MockCourierCache::new());
        let package_repository = Arc::new(MockPackageRepository::new());
        let courier_id = Uuid::new_v4();
        let order_id = Uuid::new_v4();

        repository.add_courier(create_test_courier(courier_id, CourierStatus::Free, 1));
        package_repository.add_package(create_test_package(
            order_id,
            Some(courier_id),
            PackageStatus::Assigned,
        ));

        let activities = create_activities(repository.clone(), cache.clone(), package_repository);
        let result = activities.assign_order(courier_id, order_id).await;

        assert!(result.is_ok(), "Expected Ok, got: {:?}", result.err());
        assert_eq!(repository.save_calls(), 0);
        assert_eq!(cache.cache_calls(), 0);
    }

    #[tokio::test]
    async fn assign_order_returns_courier_unavailable_when_courier_disappeared() {
        let repository = Arc::new(MockCourierRepository::new());
        let cache = Arc::new(MockCourierCache::new());
        let package_repository = Arc::new(MockPackageRepository::new());

        let activities = create_activities(repository, cache, package_repository);
        let result = activities
            .assign_order(Uuid::new_v4(), Uuid::new_v4())
            .await;

        assert!(matches!(
            result,
            Err(DeliveryActivityError::CourierUnavailable(_))
        ));
    }

    #[tokio::test]
    async fn complete_delivery_returns_ok_when_package_is_already_delivered_for_same_courier() {
        let repository = Arc::new(MockCourierRepository::new());
        let cache = Arc::new(MockCourierCache::new());
        let package_repository = Arc::new(MockPackageRepository::new());
        let courier_id = Uuid::new_v4();
        let order_id = Uuid::new_v4();

        repository.add_courier(create_test_courier(courier_id, CourierStatus::Free, 0));
        package_repository.add_package(create_test_package(
            order_id,
            Some(courier_id),
            PackageStatus::Delivered,
        ));

        let activities = create_activities(repository.clone(), cache.clone(), package_repository);
        let result = activities
            .complete_delivery(courier_id, order_id, true)
            .await;

        assert!(result.is_ok(), "Expected Ok, got: {:?}", result.err());
        assert_eq!(repository.save_calls(), 0);
        assert_eq!(cache.cache_calls(), 0);
    }

    #[tokio::test]
    async fn complete_delivery_returns_courier_unavailable_when_courier_disappeared() {
        let repository = Arc::new(MockCourierRepository::new());
        let cache = Arc::new(MockCourierCache::new());
        let package_repository = Arc::new(MockPackageRepository::new());

        let activities = create_activities(repository, cache, package_repository);
        let result = activities
            .complete_delivery(Uuid::new_v4(), Uuid::new_v4(), true)
            .await;

        assert!(matches!(
            result,
            Err(DeliveryActivityError::CourierUnavailable(_))
        ));
    }
}
