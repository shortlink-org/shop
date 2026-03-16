//! Deliver Order Handler
//!
//! Handles confirming delivery by courier.
//!
//! ## Flow
//! 1. Load package from repository
//! 2. Validate courier is assigned to package
//! 3. Update package status (DELIVERED or NOT_DELIVERED)
//! 4. Update courier stats and load
//! 5. Save package
//! 6. Publish delivery event (PackageDelivered or PackageNotDelivered)
//! 7. Return response

use std::sync::Arc;

use chrono::{DateTime, Utc};
use thiserror::Error;
use tracing::{info, warn};
use uuid::Uuid;

use crate::domain::model::courier::CourierError;
use crate::domain::model::domain::delivery::common::v1 as proto_common;
use crate::domain::model::domain::delivery::events::v1::{
    PackageDeliveredEvent, PackageNotDeliveredEvent,
};
use crate::domain::model::package::{
    NotDeliveredDetails as PackageNotDeliveredDetails, NotDeliveredReasonCode, Package, PackageId,
    PackageStatus,
};
use crate::domain::model::vo::location::Location;
use crate::domain::ports::{
    CommandHandlerWithResult, CourierCache, CourierRepository, DomainEvent, GeolocationService,
    PackageRepository, RepositoryError,
};

use super::command::{ConfirmDelivered, ConfirmNotDelivered, NotDeliveredReason};

/// Errors that can occur during delivery confirmation
#[derive(Debug, Error)]
pub enum DeliverOrderError {
    /// Package not found
    #[error("Package not found: {0}")]
    PackageNotFound(Uuid),

    /// Courier not found
    #[error("Courier not found: {0}")]
    CourierNotFound(Uuid),

    /// Courier not assigned to package
    #[error("Courier {0} is not assigned to package {1}")]
    CourierNotAssigned(Uuid, Uuid),

    /// Invalid package status
    #[error("Invalid package status for delivery: expected Assigned or InTransit, got {0}")]
    InvalidPackageStatus(PackageStatus),

    /// OTHER reason requires non-empty description
    #[error("OTHER requires description")]
    OtherReasonRequiresDescription,

    /// Package already delivered
    #[error("Package already delivered: {0}")]
    AlreadyDelivered(Uuid),

    /// Courier state transition failed
    #[error("Courier state transition failed: {0}")]
    CourierStateTransition(#[from] CourierError),

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),
}

/// Response from delivering an order
#[derive(Debug, Clone)]
pub struct Response {
    /// The package ID
    pub package_id: Uuid,
    /// The updated package status
    pub status: PackageStatus,
    /// When the package was updated
    pub updated_at: DateTime<Utc>,
}

/// Deliver Order Handler
pub struct Handler {
    courier_repo: Arc<dyn CourierRepository>,
    courier_cache: Arc<dyn CourierCache>,
    package_repo: Arc<dyn PackageRepository>,
    geolocation_service: Arc<dyn GeolocationService>,
}

enum DeliveryOutcome {
    Delivered(ConfirmDelivered),
    NotDelivered(ConfirmNotDelivered),
}

impl DeliveryOutcome {
    fn package_id(&self) -> Uuid {
        match self {
            Self::Delivered(cmd) => cmd.package_id(),
            Self::NotDelivered(cmd) => cmd.package_id(),
        }
    }

    fn courier_id(&self) -> Uuid {
        match self {
            Self::Delivered(cmd) => cmd.courier_id(),
            Self::NotDelivered(cmd) => cmd.courier_id(),
        }
    }

    fn confirmation_location(&self) -> Location {
        match self {
            Self::Delivered(cmd) => cmd.confirmation_location(),
            Self::NotDelivered(cmd) => cmd.confirmation_location(),
        }
    }

    fn log_reason(&self) -> Option<&NotDeliveredReason> {
        match self {
            Self::Delivered(_) => None,
            Self::NotDelivered(cmd) => Some(cmd.reason()),
        }
    }
}

impl Handler {
    /// Create a new handler instance
    pub fn new(
        courier_repo: Arc<dyn CourierRepository>,
        courier_cache: Arc<dyn CourierCache>,
        package_repo: Arc<dyn PackageRepository>,
        geolocation_service: Arc<dyn GeolocationService>,
    ) -> Self {
        Self {
            courier_repo,
            courier_cache,
            package_repo,
            geolocation_service,
        }
    }

    /// Convert command reason into structured package state.
    fn reason_to_package_details(reason: &NotDeliveredReason) -> PackageNotDeliveredDetails {
        match reason {
            NotDeliveredReason::CustomerUnavailable => {
                PackageNotDeliveredDetails::new(NotDeliveredReasonCode::CustomerUnavailable, None)
            }
            NotDeliveredReason::WrongAddress => {
                PackageNotDeliveredDetails::new(NotDeliveredReasonCode::WrongAddress, None)
            }
            NotDeliveredReason::Refused => {
                PackageNotDeliveredDetails::new(NotDeliveredReasonCode::Refused, None)
            }
            NotDeliveredReason::AccessDenied => {
                PackageNotDeliveredDetails::new(NotDeliveredReasonCode::AccessDenied, None)
            }
            NotDeliveredReason::Other(desc) => {
                PackageNotDeliveredDetails::new(NotDeliveredReasonCode::Other, Some(desc.clone()))
            }
        }
    }

    /// Convert NotDeliveredReason to proto enum
    fn reason_to_proto(reason: &NotDeliveredReason) -> proto_common::NotDeliveredReason {
        match reason {
            NotDeliveredReason::CustomerUnavailable => {
                proto_common::NotDeliveredReason::CustomerNotAvailable
            }
            NotDeliveredReason::WrongAddress => proto_common::NotDeliveredReason::WrongAddress,
            NotDeliveredReason::Refused => proto_common::NotDeliveredReason::CustomerRefused,
            NotDeliveredReason::AccessDenied => proto_common::NotDeliveredReason::AccessDenied,
            NotDeliveredReason::Other(_) => proto_common::NotDeliveredReason::Other,
        }
    }

    /// Get reason description for OTHER reason type
    fn get_reason_description(reason: &NotDeliveredReason) -> String {
        match reason {
            NotDeliveredReason::Other(desc) => desc.clone(),
            _ => String::new(),
        }
    }

    fn timestamp(now: DateTime<Utc>) -> pbjson_types::Timestamp {
        pbjson_types::Timestamp {
            seconds: now.timestamp(),
            nanos: now.timestamp_subsec_nanos() as i32,
        }
    }

    fn proto_location(location: &Location) -> proto_common::Location {
        proto_common::Location {
            latitude: location.latitude(),
            longitude: location.longitude(),
            accuracy: location.accuracy(),
            timestamp: None,
            speed: None,
            heading: None,
        }
    }

    fn build_delivered_event(
        package: &Package,
        cmd: &ConfirmDelivered,
        now: DateTime<Utc>,
    ) -> DomainEvent {
        DomainEvent::PackageDelivered(PackageDeliveredEvent {
            package_id: package.id().0.to_string(),
            order_id: package.order_id().to_string(),
            courier_id: cmd.courier_id().to_string(),
            status: proto_common::PackageStatus::Delivered as i32,
            delivered_at: Some(Self::timestamp(now)),
            delivery_location: Some(Self::proto_location(&cmd.confirmation_location())),
            photo: cmd
                .photo_proof()
                .map(|s| s.as_bytes().to_vec())
                .unwrap_or_default(),
            customer_signature: cmd
                .signature()
                .map(|s| s.as_bytes().to_vec())
                .unwrap_or_default(),
            occurred_at: Some(Self::timestamp(now)),
        })
    }

    fn build_not_delivered_event(
        package: &Package,
        cmd: &ConfirmNotDelivered,
        now: DateTime<Utc>,
    ) -> DomainEvent {
        let reason = Self::reason_to_proto(cmd.reason());
        let description = Self::get_reason_description(cmd.reason());

        DomainEvent::PackageNotDelivered(PackageNotDeliveredEvent {
            package_id: package.id().0.to_string(),
            order_id: package.order_id().to_string(),
            courier_id: cmd.courier_id().to_string(),
            status: proto_common::PackageStatus::NotDelivered as i32,
            not_delivered_details: Some(proto_common::NotDeliveredDetails {
                reason: reason as i32,
                description,
            }),
            not_delivered_at: Some(Self::timestamp(now)),
            courier_location: Some(Self::proto_location(&cmd.confirmation_location())),
            occurred_at: Some(Self::timestamp(now)),
        })
    }

    async fn execute_delivery(
        &self,
        outcome: DeliveryOutcome,
    ) -> Result<Response, DeliverOrderError> {
        let package_id = outcome.package_id();
        let courier_id = outcome.courier_id();

        // 1. Load package from repository
        let mut package = self
            .package_repo
            .find_by_id(PackageId::from_uuid(package_id))
            .await?
            .ok_or(DeliverOrderError::PackageNotFound(package_id))?;

        // 2. Validate package status (must be Assigned or InTransit)
        match package.status() {
            PackageStatus::Assigned | PackageStatus::InTransit => {}
            PackageStatus::Delivered => {
                return Err(DeliverOrderError::AlreadyDelivered(package_id));
            }
            status => {
                return Err(DeliverOrderError::InvalidPackageStatus(status));
            }
        }

        // 3. Validate courier is assigned to this package
        let assigned_courier =
            package
                .courier_id()
                .ok_or(DeliverOrderError::CourierNotAssigned(
                    courier_id, package_id,
                ))?;

        if assigned_courier != courier_id {
            return Err(DeliverOrderError::CourierNotAssigned(
                courier_id, package_id,
            ));
        }

        // 4. Load courier aggregate
        let mut courier = self
            .courier_repo
            .find_by_id(courier_id)
            .await?
            .ok_or(DeliverOrderError::CourierNotFound(courier_id))?;

        let now = Utc::now();
        let events = vec![match &outcome {
            DeliveryOutcome::Delivered(cmd) => {
                package
                    .mark_delivered()
                    .map_err(|_| DeliverOrderError::InvalidPackageStatus(package.status()))?;
                courier.complete_delivery(true)?;
                Self::build_delivered_event(&package, cmd, now)
            }
            DeliveryOutcome::NotDelivered(cmd) => {
                package
                    .mark_not_delivered(Self::reason_to_package_details(cmd.reason()))
                    .map_err(|_| DeliverOrderError::InvalidPackageStatus(package.status()))?;
                courier.complete_delivery(false)?;
                Self::build_not_delivered_event(&package, cmd, now)
            }
        }];

        // 5. Save both aggregates and outbox rows atomically.
        self.package_repo
            .save_courier_with_package_and_events(&courier, &package, &events)
            .await?;
        self.courier_cache.cache(&courier).await.map_err(|e| {
            RepositoryError::QueryError(format!("Failed to refresh courier cache: {}", e))
        })?;
        info!(
            package_id = %package_id,
            courier_id = %courier_id,
            reason = ?outcome.log_reason(),
            "Delivery lifecycle event enqueued to outbox"
        );

        // 6. OMS is notified later by outbox forwarder.
        // 7. Update courier location via GeolocationService using the same command timestamp.
        if let Err(e) = self
            .geolocation_service
            .update_location(courier_id, outcome.confirmation_location(), now)
            .await
        {
            warn!(
                package_id = %package_id,
                courier_id = %courier_id,
                error = %e,
                "Failed to update courier location via GeolocationService (non-fatal)"
            );
        }

        Ok(Response {
            package_id,
            status: package.status(),
            updated_at: package.updated_at(),
        })
    }
}

impl CommandHandlerWithResult<ConfirmDelivered, Response> for Handler {
    type Error = DeliverOrderError;

    async fn handle(&self, cmd: ConfirmDelivered) -> Result<Response, Self::Error> {
        self.execute_delivery(DeliveryOutcome::Delivered(cmd)).await
    }
}

impl CommandHandlerWithResult<ConfirmNotDelivered, Response> for Handler {
    type Error = DeliverOrderError;

    async fn handle(&self, cmd: ConfirmNotDelivered) -> Result<Response, Self::Error> {
        if let NotDeliveredReason::Other(desc) = cmd.reason() {
            if desc.trim().is_empty() {
                return Err(DeliverOrderError::OtherReasonRequiresDescription);
            }
        }

        self.execute_delivery(DeliveryOutcome::NotDelivered(cmd))
            .await
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::courier::{Courier, CourierId, WorkHours};
    use crate::domain::model::package::{Address, DeliveryPeriod, Package, Priority};
    use crate::domain::model::vo::location::Location;
    use crate::domain::model::vo::TransportType;
    use crate::domain::ports::{
        CacheError, CachedCourierState, GeolocationService, GeolocationServiceError, PackageFilter,
    };
    use async_trait::async_trait;
    use chrono::{NaiveTime, Utc};
    use std::collections::HashMap;
    use std::sync::Mutex;

    // ==================== Mock Repositories ====================

    struct MockCourierRepository {
        couriers: Mutex<HashMap<Uuid, Courier>>,
    }

    impl MockCourierRepository {
        fn new() -> Self {
            Self {
                couriers: Mutex::new(HashMap::new()),
            }
        }

        fn add_courier(&self, courier: Courier) {
            let mut couriers = self.couriers.lock().unwrap();
            couriers.insert(courier.id().0, courier);
        }
    }

    #[async_trait]
    impl CourierRepository for MockCourierRepository {
        async fn save(&self, courier: &Courier) -> Result<(), RepositoryError> {
            let mut couriers = self.couriers.lock().unwrap();
            couriers.insert(courier.id().0, courier.clone());
            Ok(())
        }

        async fn find_by_id(&self, id: Uuid) -> Result<Option<Courier>, RepositoryError> {
            let couriers = self.couriers.lock().unwrap();
            Ok(couriers.get(&id).cloned())
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
            let couriers = self.couriers.lock().unwrap();
            Ok(couriers.values().cloned().collect())
        }

        async fn find_by_filter(
            &self,
            _filter: crate::domain::ports::CourierFilter,
            _limit: u64,
            _offset: u64,
        ) -> Result<Vec<Courier>, RepositoryError> {
            let couriers = self.couriers.lock().unwrap();
            Ok(couriers.values().cloned().collect())
        }

        async fn count_by_filter(
            &self,
            _filter: crate::domain::ports::CourierFilter,
        ) -> Result<u64, RepositoryError> {
            let couriers = self.couriers.lock().unwrap();
            Ok(couriers.len() as u64)
        }
    }

    struct MockCourierCache {
        fail_cache: bool,
        cache_calls: Mutex<u32>,
        cached_couriers: Mutex<HashMap<Uuid, Courier>>,
    }

    impl MockCourierCache {
        fn new() -> Self {
            Self {
                fail_cache: false,
                cache_calls: Mutex::new(0),
                cached_couriers: Mutex::new(HashMap::new()),
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
            self.cached_couriers
                .lock()
                .unwrap()
                .get(&courier_id)
                .cloned()
        }
    }

    #[async_trait]
    impl CourierCache for MockCourierCache {
        async fn cache(&self, courier: &Courier) -> Result<(), CacheError> {
            *self.cache_calls.lock().unwrap() += 1;
            if self.fail_cache {
                return Err(CacheError::OperationError(
                    "simulated cache refresh failure".to_string(),
                ));
            }

            self.cached_couriers
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

        async fn get_status(
            &self,
            _courier_id: Uuid,
        ) -> Result<Option<crate::domain::model::courier::CourierStatus>, CacheError> {
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
        saved_couriers: Mutex<HashMap<Uuid, Courier>>,
        fail_transactional_save: bool,
        transactional_save_calls: Mutex<u32>,
    }

    impl MockPackageRepository {
        fn new() -> Self {
            Self {
                packages: Mutex::new(HashMap::new()),
                saved_couriers: Mutex::new(HashMap::new()),
                fail_transactional_save: false,
                transactional_save_calls: Mutex::new(0),
            }
        }

        fn failing_on_transactional_save() -> Self {
            Self {
                fail_transactional_save: true,
                ..Self::new()
            }
        }

        fn add_package(&self, package: Package) {
            let mut packages = self.packages.lock().unwrap();
            packages.insert(package.id().0, package);
        }

        fn saved_courier(&self, courier_id: Uuid) -> Option<Courier> {
            self.saved_couriers
                .lock()
                .unwrap()
                .get(&courier_id)
                .cloned()
        }

        fn transactional_save_calls(&self) -> u32 {
            *self.transactional_save_calls.lock().unwrap()
        }
    }

    #[async_trait]
    impl PackageRepository for MockPackageRepository {
        async fn save(&self, package: &Package) -> Result<(), RepositoryError> {
            let mut packages = self.packages.lock().unwrap();
            packages.insert(package.id().0, package.clone());
            Ok(())
        }

        async fn save_courier_with_package_and_events(
            &self,
            courier: &Courier,
            package: &Package,
            _events: &[DomainEvent],
        ) -> Result<(), RepositoryError> {
            *self.transactional_save_calls.lock().unwrap() += 1;
            if self.fail_transactional_save {
                return Err(RepositoryError::QueryError(
                    "simulated transactional save failure".to_string(),
                ));
            }

            self.saved_couriers
                .lock()
                .unwrap()
                .insert(courier.id().0, courier.clone());
            self.save(package).await
        }

        async fn find_by_id(&self, id: PackageId) -> Result<Option<Package>, RepositoryError> {
            let packages = self.packages.lock().unwrap();
            Ok(packages.get(&id.0).cloned())
        }

        async fn find_by_order_id(
            &self,
            _order_id: Uuid,
        ) -> Result<Option<Package>, RepositoryError> {
            Ok(None)
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

    struct MockGeolocationService {
        fail_update: bool,
        update_calls: Mutex<u32>,
    }

    impl MockGeolocationService {
        fn new() -> Self {
            Self {
                fail_update: false,
                update_calls: Mutex::new(0),
            }
        }

        fn failing_on_update() -> Self {
            Self {
                fail_update: true,
                ..Self::new()
            }
        }

        fn update_calls(&self) -> u32 {
            *self.update_calls.lock().unwrap()
        }
    }

    #[async_trait]
    impl GeolocationService for MockGeolocationService {
        async fn update_location(
            &self,
            _courier_id: Uuid,
            _location: Location,
            _timestamp: chrono::DateTime<Utc>,
        ) -> Result<(), GeolocationServiceError> {
            *self.update_calls.lock().unwrap() += 1;
            if self.fail_update {
                return Err(GeolocationServiceError::RepositoryError(
                    "simulated geolocation failure".to_string(),
                ));
            }
            Ok(())
        }

        async fn get_location(
            &self,
            _courier_id: Uuid,
        ) -> Result<Option<crate::domain::model::CourierLocation>, GeolocationServiceError>
        {
            Ok(None)
        }
    }

    // ==================== Test Helpers ====================

    fn create_test_address() -> Address {
        Address::new(
            "123 Main St".to_string(),
            "Berlin".to_string(),
            Location::new(52.52, 13.405, 10.0).unwrap(),
        )
    }

    fn create_test_location() -> Location {
        Location::new(52.52, 13.405, 10.0).unwrap()
    }

    fn create_in_transit_package(courier_id: Uuid) -> Package {
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
            "Berlin-101".to_string(),
        );

        // Move to pool -> assign -> start transit
        package.move_to_pool().unwrap();
        package.assign_to(courier_id).unwrap();
        package.start_transit().unwrap();
        package
    }

    fn create_assigned_package(courier_id: Uuid) -> Package {
        let now = Utc::now();
        let period = DeliveryPeriod::new(
            now + chrono::Duration::hours(2),
            now + chrono::Duration::hours(4),
        )
        .unwrap();

        let mut package = Package::new(
            Uuid::new_v4(),
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
        package.assign_to(courier_id).unwrap();
        package
    }

    fn create_delivered_package(courier_id: Uuid) -> Package {
        let mut package = create_in_transit_package(courier_id);
        package.mark_delivered().unwrap();
        package
    }

    fn create_test_courier(id: Uuid) -> Courier {
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
            crate::domain::model::courier::CourierStatus::Free,
            1,
            4.5,
            10, // successful_deliveries
            1,  // failed_deliveries
            Utc::now(),
            Utc::now(),
            1,
        )
        .unwrap()
    }

    // ==================== Tests ====================

    #[tokio::test]
    async fn test_deliver_order_success() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::new());

        let courier_id = Uuid::new_v4();
        let courier = create_test_courier(courier_id);
        courier_repo.add_courier(courier);

        let package = create_in_transit_package(courier_id);
        let package_id = package.id().0;
        package_repo.add_package(package);

        let handler = Handler::new(
            courier_repo,
            courier_cache,
            package_repo.clone(),
            Arc::new(MockGeolocationService::new()),
        );

        let cmd = ConfirmDelivered::new(package_id, courier_id, create_test_location(), None, None);
        let result = handler.handle(cmd).await;

        assert!(result.is_ok(), "Expected Ok, got: {:?}", result.err());
        let response = result.unwrap();
        assert_eq!(response.package_id, package_id);
        assert_eq!(response.status, PackageStatus::Delivered);

        // Verify package status changed
        let updated_package = package_repo
            .find_by_id(PackageId::from_uuid(package_id))
            .await
            .unwrap();
        assert!(updated_package.is_some());
        assert_eq!(updated_package.unwrap().status(), PackageStatus::Delivered);
    }

    #[tokio::test]
    async fn test_deliver_order_not_delivered_success() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::new());

        let courier_id = Uuid::new_v4();
        let courier = create_test_courier(courier_id);
        courier_repo.add_courier(courier);

        let package = create_in_transit_package(courier_id);
        let package_id = package.id().0;
        package_repo.add_package(package);

        let handler = Handler::new(
            courier_repo,
            courier_cache,
            package_repo.clone(),
            Arc::new(MockGeolocationService::new()),
        );

        let cmd = ConfirmNotDelivered::new(
            package_id,
            courier_id,
            create_test_location(),
            NotDeliveredReason::CustomerUnavailable,
        );
        let result = handler.handle(cmd).await;

        assert!(result.is_ok(), "Expected Ok, got: {:?}", result.err());
        let response = result.unwrap();
        assert_eq!(response.package_id, package_id);
        assert_eq!(response.status, PackageStatus::NotDelivered);
    }

    #[tokio::test]
    async fn test_deliver_order_package_not_found() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::new());

        let handler = Handler::new(
            courier_repo,
            courier_cache,
            package_repo,
            Arc::new(MockGeolocationService::new()),
        );

        let cmd = ConfirmDelivered::new(
            Uuid::new_v4(),
            Uuid::new_v4(),
            create_test_location(),
            None,
            None,
        );
        let result = handler.handle(cmd).await;

        assert!(matches!(result, Err(DeliverOrderError::PackageNotFound(_))));
    }

    #[tokio::test]
    async fn test_deliver_order_courier_not_assigned() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::new());

        let assigned_courier_id = Uuid::new_v4();
        let wrong_courier_id = Uuid::new_v4();

        // Add both couriers
        let assigned_courier = create_test_courier(assigned_courier_id);
        let wrong_courier = create_test_courier(wrong_courier_id);
        courier_repo.add_courier(assigned_courier);
        courier_repo.add_courier(wrong_courier);

        // Package is assigned to assigned_courier_id
        let package = create_in_transit_package(assigned_courier_id);
        let package_id = package.id().0;
        package_repo.add_package(package);

        let handler = Handler::new(
            courier_repo,
            courier_cache,
            package_repo,
            Arc::new(MockGeolocationService::new()),
        );

        // Try to deliver with wrong courier
        let cmd = ConfirmDelivered::new(
            package_id,
            wrong_courier_id,
            create_test_location(),
            None,
            None,
        );
        let result = handler.handle(cmd).await;

        assert!(matches!(
            result,
            Err(DeliverOrderError::CourierNotAssigned(_, _))
        ));
    }

    #[tokio::test]
    async fn test_other_reason_requires_description() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::new());
        let handler = Handler::new(
            courier_repo,
            courier_cache,
            package_repo,
            Arc::new(MockGeolocationService::new()),
        );

        let cmd = ConfirmNotDelivered::new(
            Uuid::new_v4(),
            Uuid::new_v4(),
            create_test_location(),
            NotDeliveredReason::Other(String::new()), // empty description
        );
        let result = handler.handle(cmd).await;
        assert!(matches!(
            result,
            Err(DeliverOrderError::OtherReasonRequiresDescription)
        ));
    }

    #[tokio::test]
    async fn test_deliver_order_invalid_status() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::new());

        let courier_id = Uuid::new_v4();
        let courier = create_test_courier(courier_id);
        courier_repo.add_courier(courier);

        // Create package in InPool status (not Assigned)
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
            "Berlin-101".to_string(),
        );
        package.move_to_pool().unwrap();
        // NOT assigned - status is InPool
        let package_id = package.id().0;
        package_repo.add_package(package);

        let handler = Handler::new(
            courier_repo,
            courier_cache,
            package_repo,
            Arc::new(MockGeolocationService::new()),
        );

        let cmd = ConfirmDelivered::new(package_id, courier_id, create_test_location(), None, None);
        let result = handler.handle(cmd).await;

        assert!(matches!(
            result,
            Err(DeliverOrderError::InvalidPackageStatus(_))
        ));
    }

    #[tokio::test]
    async fn test_deliver_order_already_delivered() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::new());

        let courier_id = Uuid::new_v4();
        courier_repo.add_courier(create_test_courier(courier_id));

        let package = create_delivered_package(courier_id);
        let package_id = package.id().0;
        package_repo.add_package(package);

        let handler = Handler::new(
            courier_repo,
            courier_cache,
            package_repo,
            Arc::new(MockGeolocationService::new()),
        );

        let result = handler
            .handle(ConfirmDelivered::new(
                package_id,
                courier_id,
                create_test_location(),
                None,
                None,
            ))
            .await;

        assert!(matches!(
            result,
            Err(DeliverOrderError::AlreadyDelivered(id)) if id == package_id
        ));
    }

    #[tokio::test]
    async fn test_deliver_order_courier_not_found() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::new());

        let courier_id = Uuid::new_v4();
        let package = create_in_transit_package(courier_id);
        let package_id = package.id().0;
        package_repo.add_package(package);

        let handler = Handler::new(
            courier_repo,
            courier_cache,
            package_repo,
            Arc::new(MockGeolocationService::new()),
        );

        let result = handler
            .handle(ConfirmDelivered::new(
                package_id,
                courier_id,
                create_test_location(),
                None,
                None,
            ))
            .await;

        assert!(matches!(
            result,
            Err(DeliverOrderError::CourierNotFound(id)) if id == courier_id
        ));
    }

    #[tokio::test]
    async fn test_deliver_order_assigned_status_fails_on_domain_transition() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::new());

        let courier_id = Uuid::new_v4();
        courier_repo.add_courier(create_test_courier(courier_id));

        let package = create_assigned_package(courier_id);
        let package_id = package.id().0;
        package_repo.add_package(package);

        let handler = Handler::new(
            courier_repo,
            courier_cache,
            package_repo,
            Arc::new(MockGeolocationService::new()),
        );

        let result = handler
            .handle(ConfirmDelivered::new(
                package_id,
                courier_id,
                create_test_location(),
                None,
                None,
            ))
            .await;

        assert!(matches!(
            result,
            Err(DeliverOrderError::InvalidPackageStatus(
                PackageStatus::Assigned
            ))
        ));
    }

    #[tokio::test]
    async fn test_deliver_order_cache_failure_returns_error_after_persistence() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::failing_on_cache());
        let package_repo = Arc::new(MockPackageRepository::new());
        let geolocation_service = Arc::new(MockGeolocationService::new());

        let courier_id = Uuid::new_v4();
        courier_repo.add_courier(create_test_courier(courier_id));

        let package = create_in_transit_package(courier_id);
        let package_id = package.id().0;
        package_repo.add_package(package);

        let handler = Handler::new(
            courier_repo,
            courier_cache.clone(),
            package_repo.clone(),
            geolocation_service.clone(),
        );

        let result = handler
            .handle(ConfirmDelivered::new(
                package_id,
                courier_id,
                create_test_location(),
                None,
                None,
            ))
            .await;

        assert!(matches!(
            result,
            Err(DeliverOrderError::RepositoryError(
                RepositoryError::QueryError(_)
            ))
        ));
        assert_eq!(package_repo.transactional_save_calls(), 1);
        assert!(package_repo.saved_courier(courier_id).is_some());
        assert_eq!(courier_cache.cache_calls(), 1);
        assert!(courier_cache.cached_courier(courier_id).is_none());
        assert_eq!(geolocation_service.update_calls(), 0);

        let persisted = package_repo
            .find_by_id(PackageId::from_uuid(package_id))
            .await
            .unwrap()
            .unwrap();
        assert_eq!(persisted.status(), PackageStatus::Delivered);
    }

    #[tokio::test]
    async fn test_deliver_order_geolocation_failure_is_non_fatal() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::new());
        let geolocation_service = Arc::new(MockGeolocationService::failing_on_update());

        let courier_id = Uuid::new_v4();
        courier_repo.add_courier(create_test_courier(courier_id));

        let package = create_in_transit_package(courier_id);
        let package_id = package.id().0;
        package_repo.add_package(package);

        let handler = Handler::new(
            courier_repo,
            courier_cache.clone(),
            package_repo.clone(),
            geolocation_service.clone(),
        );

        let result = handler
            .handle(ConfirmDelivered::new(
                package_id,
                courier_id,
                create_test_location(),
                None,
                None,
            ))
            .await;

        assert!(result.is_ok(), "Expected Ok, got: {:?}", result.err());
        assert_eq!(courier_cache.cache_calls(), 1);
        assert!(courier_cache.cached_courier(courier_id).is_some());
        assert_eq!(geolocation_service.update_calls(), 1);

        let persisted = package_repo
            .find_by_id(PackageId::from_uuid(package_id))
            .await
            .unwrap()
            .unwrap();
        assert_eq!(persisted.status(), PackageStatus::Delivered);
    }

    #[tokio::test]
    async fn test_deliver_order_transactional_save_failure_keeps_state_unpersisted() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::failing_on_transactional_save());
        let geolocation_service = Arc::new(MockGeolocationService::new());

        let courier_id = Uuid::new_v4();
        courier_repo.add_courier(create_test_courier(courier_id));

        let package = create_in_transit_package(courier_id);
        let package_id = package.id().0;
        package_repo.add_package(package);

        let handler = Handler::new(
            courier_repo,
            courier_cache.clone(),
            package_repo.clone(),
            geolocation_service.clone(),
        );

        let result = handler
            .handle(ConfirmDelivered::new(
                package_id,
                courier_id,
                create_test_location(),
                None,
                None,
            ))
            .await;

        assert!(matches!(
            result,
            Err(DeliverOrderError::RepositoryError(
                RepositoryError::QueryError(_)
            ))
        ));
        assert_eq!(package_repo.transactional_save_calls(), 1);
        assert!(package_repo.saved_courier(courier_id).is_none());
        assert_eq!(courier_cache.cache_calls(), 0);
        assert_eq!(geolocation_service.update_calls(), 0);

        let persisted = package_repo
            .find_by_id(PackageId::from_uuid(package_id))
            .await
            .unwrap()
            .unwrap();
        assert_eq!(persisted.status(), PackageStatus::InTransit);
    }
}
