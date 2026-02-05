//! Assign Order Handler
//!
//! Handles assigning a package to a courier.
//!
//! ## Flow
//! 1. Load package from repository
//! 2. Validate package status is InPool
//! 3. If auto-assign: use DispatchService to find nearest courier
//! 4. If manual: validate assignment using AssignmentValidationService
//! 5. Update package status to ASSIGNED
//! 6. Update courier load
//! 7. Save package to repository
//! 8. Publish PackageAssigned event
//! 9. Send push notification to courier

use std::sync::Arc;

use chrono::{Timelike, Utc};
use thiserror::Error;
use tracing::{info, warn};
use uuid::Uuid;

use crate::domain::model::courier::CourierStatus;
use crate::domain::model::domain::delivery::common::v1 as proto_common;
use crate::domain::model::domain::delivery::events::v1::PackageAssignedEvent;
use crate::domain::model::package::{PackageId, PackageStatus};
use crate::domain::ports::{
    CommandHandlerWithResult, CourierCache, CourierRepository, DomainEvent, EventPublisher,
    LocationCache, NotificationService, OrderAssignedNotification, PackageRepository,
    RepositoryError,
};
use crate::domain::services::assignment_validation::{
    AssignmentValidationService, CourierAvailability, PackageForValidation,
};
use crate::domain::services::dispatch::{CourierForDispatch, DispatchService, PackageForDispatch};

use super::command::AssignmentMode;
use super::Command;

/// Errors that can occur during order assignment
#[derive(Debug, Error)]
pub enum AssignOrderError {
    /// Package not found
    #[error("Package not found: {0}")]
    PackageNotFound(Uuid),

    /// Courier not found
    #[error("Courier not found: {0}")]
    CourierNotFound(Uuid),

    /// No available courier
    #[error("No available courier for package in zone: {0}")]
    NoAvailableCourier(String),

    /// Courier ID required for manual assignment
    #[error("Courier ID required for manual assignment")]
    CourierIdRequired,

    /// Courier not available
    #[error("Courier not available: {0}")]
    CourierNotAvailable(String),

    /// Assignment validation failed
    #[error("Assignment validation failed: {0}")]
    ValidationFailed(String),

    /// Invalid package status for assignment
    #[error("Invalid package status: expected InPool, got {0}")]
    InvalidPackageStatus(PackageStatus),

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    /// Event publishing error (non-fatal, logged as warning)
    #[error("Event publishing error: {0}")]
    EventPublishError(String),

    /// Notification error (non-fatal, logged as warning)
    #[error("Notification error: {0}")]
    NotificationError(String),
}

/// Response from assigning an order
#[derive(Debug, Clone)]
pub struct Response {
    /// The package ID
    pub package_id: Uuid,
    /// The assigned courier ID
    pub courier_id: Uuid,
    /// Estimated delivery time in minutes
    pub estimated_delivery_minutes: u32,
}

/// Assign Order Handler
pub struct Handler<R, C, P, E, N, L>
where
    R: CourierRepository,
    C: CourierCache,
    P: PackageRepository,
    E: EventPublisher,
    N: NotificationService,
    L: LocationCache,
{
    courier_repo: Arc<R>,
    courier_cache: Arc<C>,
    package_repo: Arc<P>,
    event_publisher: Arc<E>,
    notification_service: Arc<N>,
    location_cache: Arc<L>,
}

impl<R, C, P, E, N, L> Handler<R, C, P, E, N, L>
where
    R: CourierRepository,
    C: CourierCache,
    P: PackageRepository,
    E: EventPublisher,
    N: NotificationService,
    L: LocationCache,
{
    /// Create a new handler instance
    pub fn new(
        courier_repo: Arc<R>,
        courier_cache: Arc<C>,
        package_repo: Arc<P>,
        event_publisher: Arc<E>,
        notification_service: Arc<N>,
        location_cache: Arc<L>,
    ) -> Self {
        Self {
            courier_repo,
            courier_cache,
            package_repo,
            event_publisher,
            notification_service,
            location_cache,
        }
    }

    /// Convert Courier entity to CourierForDispatch with location from cache
    async fn courier_to_dispatch_with_location(
        &self,
        courier: &crate::domain::model::courier::Courier,
    ) -> CourierForDispatch {
        // Try to get location from cache
        let current_location = match self.location_cache.get_location(courier.id().0).await {
            Ok(Some(loc)) => Some(loc.location().clone()),
            Ok(None) => None,
            Err(e) => {
                warn!(
                    courier_id = %courier.id().0,
                    error = %e,
                    "Failed to get courier location from cache"
                );
                None
            }
        };

        CourierForDispatch {
            id: courier.id().0.to_string(),
            status: courier.status(),
            transport_type: courier.transport_type(),
            max_distance_km: courier.max_distance_km(),
            capacity: courier.capacity().clone(),
            current_location,
            rating: courier.rating(),
            work_zone: courier.work_zone().to_string(),
        }
    }

    /// Get current hour for working hours validation
    fn current_hour() -> u8 {
        Utc::now().time().hour() as u8
    }
}

impl<R, C, P, E, N, L> CommandHandlerWithResult<Command, Response> for Handler<R, C, P, E, N, L>
where
    R: CourierRepository + Send + Sync,
    C: CourierCache + Send + Sync,
    P: PackageRepository + Send + Sync,
    E: EventPublisher + Send + Sync,
    N: NotificationService + Send + Sync,
    L: LocationCache + Send + Sync,
{
    type Error = AssignOrderError;

    /// Handle the AssignOrder command
    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        // 1. Load package from repository
        let mut package = self
            .package_repo
            .find_by_id(PackageId::from_uuid(cmd.package_id))
            .await?
            .ok_or(AssignOrderError::PackageNotFound(cmd.package_id))?;

        // 2. Validate package status is InPool
        if package.status() != PackageStatus::InPool {
            return Err(AssignOrderError::InvalidPackageStatus(package.status()));
        }

        // 3. Find courier based on assignment mode
        let (courier_id, estimated_minutes) = match cmd.mode {
            AssignmentMode::Auto => {
                // Get available couriers from cache
                let zone = package.zone();
                let free_courier_ids = self
                    .courier_cache
                    .get_free_couriers_in_zone(zone)
                    .await
                    .map_err(|e| {
                        AssignOrderError::RepositoryError(RepositoryError::QueryError(e.to_string()))
                    })?;

                if free_courier_ids.is_empty() {
                    return Err(AssignOrderError::NoAvailableCourier(zone.to_string()));
                }

                // Load courier details with locations from cache
                let mut couriers_for_dispatch = Vec::new();
                for id in free_courier_ids {
                    if let Some(courier) = self.courier_repo.find_by_id(id).await? {
                        couriers_for_dispatch.push(self.courier_to_dispatch_with_location(&courier).await);
                    }
                }

                // Create PackageForDispatch
                let package_for_dispatch = PackageForDispatch {
                    id: package.id().0.to_string(),
                    pickup_location: package.pickup_address().location.clone(),
                    delivery_zone: zone.to_string(),
                    is_urgent: package.priority() == crate::domain::model::package::Priority::Urgent,
                };

                // Use DispatchService to find nearest courier
                let dispatch_result = DispatchService::find_nearest_courier(
                    &couriers_for_dispatch,
                    &package_for_dispatch,
                )
                .map_err(|failure| {
                    tracing::debug!(
                        zone = %zone,
                        rejections = ?failure.rejections,
                        "No courier available for dispatch"
                    );
                    AssignOrderError::NoAvailableCourier(zone.to_string())
                })?;

                let courier_uuid = Uuid::parse_str(&dispatch_result.courier_id).map_err(|_| {
                    AssignOrderError::RepositoryError(RepositoryError::QueryError(
                        "Invalid courier ID from dispatch".to_string(),
                    ))
                })?;

                (courier_uuid, dispatch_result.estimated_arrival_minutes as u32)
            }
            AssignmentMode::Manual => {
                let courier_id = cmd
                    .courier_id
                    .ok_or(AssignOrderError::CourierIdRequired)?;

                // Load courier
                let courier = self
                    .courier_repo
                    .find_by_id(courier_id)
                    .await?
                    .ok_or(AssignOrderError::CourierNotFound(courier_id))?;

                // Check courier status
                if courier.status() != CourierStatus::Free {
                    return Err(AssignOrderError::CourierNotAvailable(format!(
                        "Courier status is {:?}",
                        courier.status()
                    )));
                }

                // Get distance from location cache for validation
                let distance_to_courier = match self.location_cache.get_location(courier_id).await {
                    Ok(Some(loc)) => loc.location().distance_to(&package.pickup_address().location),
                    _ => 0.0, // Default to 0 if location not available
                };

                // Validate using AssignmentValidationService
                let work_hours = courier.work_hours();
                let courier_availability = CourierAvailability {
                    status: courier.status(),
                    current_load: courier.current_load(),
                    max_load: courier.max_load(),
                    work_start_hour: work_hours.start.hour() as u8,
                    work_end_hour: work_hours.end.hour() as u8,
                    max_distance_km: courier.max_distance_km(),
                };

                let package_for_validation = PackageForValidation {
                    status: package.status(),
                    distance_to_courier_km: distance_to_courier,
                };

                AssignmentValidationService::validate(
                    &courier_availability,
                    &package_for_validation,
                    Self::current_hour(),
                )
                .map_err(|errors| {
                    let error_msgs: Vec<String> = errors.iter().map(|e| e.to_string()).collect();
                    AssignOrderError::ValidationFailed(error_msgs.join("; "))
                })?;

                // Estimate delivery time based on transport type
                let estimated = courier.transport_type().calculate_travel_time_minutes(
                    if distance_to_courier > 0.0 { distance_to_courier } else { 5.0 }
                );
                (courier_id, estimated as u32)
            }
        };

        // 4. Assign package to courier (transitions to ASSIGNED status)
        package.assign_to(courier_id).map_err(|e| {
            AssignOrderError::ValidationFailed(format!("Failed to assign package: {}", e))
        })?;

        // 5. Update courier load in cache
        let courier = self
            .courier_repo
            .find_by_id(courier_id)
            .await?
            .ok_or(AssignOrderError::CourierNotFound(courier_id))?;

        self.courier_cache
            .update_load(courier_id, courier.current_load() + 1, courier.max_load())
            .await
            .map_err(|e| {
                AssignOrderError::RepositoryError(RepositoryError::QueryError(e.to_string()))
            })?;

        // 6. Save package to repository
        self.package_repo.save(&package).await?;

        // 7. Publish PackageAssigned event
        let now = Utc::now();
        let event = PackageAssignedEvent {
            package_id: package.id().0.to_string(),
            courier_id: courier_id.to_string(),
            status: proto_common::PackageStatus::Assigned as i32,
            assigned_at: Some(pbjson_types::Timestamp {
                seconds: now.timestamp(),
                nanos: now.timestamp_subsec_nanos() as i32,
            }),
            pickup_address: Some(proto_common::Address {
                street: package.pickup_address().street.clone(),
                city: package.pickup_address().city.clone(),
                postal_code: package.pickup_address().postal_code.clone(),
                country: "DE".to_string(), // TODO: Add country to Address
                latitude: package.pickup_address().location.latitude(),
                longitude: package.pickup_address().location.longitude(),
            }),
            delivery_address: Some(proto_common::Address {
                street: package.delivery_address().street.clone(),
                city: package.delivery_address().city.clone(),
                postal_code: package.delivery_address().postal_code.clone(),
                country: "DE".to_string(),
                latitude: package.delivery_address().location.latitude(),
                longitude: package.delivery_address().location.longitude(),
            }),
            delivery_period: Some(proto_common::DeliveryPeriod {
                start_time: Some(pbjson_types::Timestamp {
                    seconds: package.delivery_period().start().timestamp(),
                    nanos: 0,
                }),
                end_time: Some(pbjson_types::Timestamp {
                    seconds: package.delivery_period().end().timestamp(),
                    nanos: 0,
                }),
            }),
            customer_phone: package.customer_phone().unwrap_or_default().to_string(),
            occurred_at: Some(pbjson_types::Timestamp {
                seconds: now.timestamp(),
                nanos: now.timestamp_subsec_nanos() as i32,
            }),
        };

        if let Err(e) = self
            .event_publisher
            .publish(DomainEvent::PackageAssigned(event))
            .await
        {
            warn!(
                package_id = %cmd.package_id,
                courier_id = %courier_id,
                error = %e,
                "Failed to publish PackageAssigned event (non-fatal)"
            );
        } else {
            info!(
                package_id = %cmd.package_id,
                courier_id = %courier_id,
                "PackageAssigned event published"
            );
        }

        // 8. Send push notification to courier
        if let Some(push_token) = courier.push_token() {
            let notification = OrderAssignedNotification {
                package_id: package.id().0,
                pickup_address: package.pickup_address().street.clone(),
                pickup_location: package.pickup_address().location.clone(),
                delivery_address: package.delivery_address().street.clone(),
                delivery_location: package.delivery_address().location.clone(),
                customer_phone: package.customer_phone().unwrap_or_default().to_string(),
                delivery_start: package.delivery_period().start().to_rfc3339(),
                delivery_end: package.delivery_period().end().to_rfc3339(),
            };

            if let Err(e) = self.notification_service.send_order_assigned(push_token, notification).await {
                warn!(
                    package_id = %cmd.package_id,
                    courier_id = %courier_id,
                    error = %e,
                    "Failed to send push notification (non-fatal)"
                );
            } else {
                info!(
                    package_id = %cmd.package_id,
                    courier_id = %courier_id,
                    "Push notification sent to courier"
                );
            }
        }

        Ok(Response {
            package_id: cmd.package_id,
            courier_id,
            estimated_delivery_minutes: estimated_minutes,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::courier::{Courier, CourierId};
    use crate::domain::model::package::{Address, DeliveryPeriod, Package, Priority};
    use crate::domain::model::vo::location::Location;
    use crate::domain::model::vo::TransportType;
    use crate::domain::model::CourierLocation;
    use crate::domain::ports::{
        CacheError, CachedCourierState, DeliveryStatusNotification, LocationCacheError,
        MockEventPublisher, NotificationError, PackageFilter,
    };
    use async_trait::async_trait;
    use chrono::Utc;
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
        free_couriers: Mutex<Vec<Uuid>>,
    }

    impl MockCourierCache {
        fn new() -> Self {
            Self {
                free_couriers: Mutex::new(Vec::new()),
            }
        }

        fn add_free_courier(&self, id: Uuid) {
            let mut couriers = self.free_couriers.lock().unwrap();
            couriers.push(id);
        }
    }

    #[async_trait]
    impl CourierCache for MockCourierCache {
        async fn initialize_state(
            &self,
            _courier_id: Uuid,
            _state: CachedCourierState,
            _work_zone: &str,
        ) -> Result<(), CacheError> {
            Ok(())
        }

        async fn get_state(&self, _courier_id: Uuid) -> Result<Option<CachedCourierState>, CacheError> {
            Ok(None)
        }

        async fn set_status(
            &self,
            _courier_id: Uuid,
            _status: CourierStatus,
            _work_zone: &str,
        ) -> Result<(), CacheError> {
            Ok(())
        }

        async fn get_status(&self, _courier_id: Uuid) -> Result<Option<CourierStatus>, CacheError> {
            Ok(Some(CourierStatus::Free))
        }

        async fn update_load(
            &self,
            _courier_id: Uuid,
            _current_load: u32,
            _max_load: u32,
        ) -> Result<(), CacheError> {
            Ok(())
        }

        async fn update_stats(
            &self,
            _courier_id: Uuid,
            _rating: f64,
            _successful_deliveries: u32,
            _failed_deliveries: u32,
        ) -> Result<(), CacheError> {
            Ok(())
        }

        async fn get_free_couriers_in_zone(&self, _zone: &str) -> Result<Vec<Uuid>, CacheError> {
            let couriers = self.free_couriers.lock().unwrap();
            Ok(couriers.clone())
        }

        async fn get_all_free_couriers(&self) -> Result<Vec<Uuid>, CacheError> {
            let couriers = self.free_couriers.lock().unwrap();
            Ok(couriers.clone())
        }

        async fn remove(&self, _courier_id: Uuid, _work_zone: &str) -> Result<(), CacheError> {
            Ok(())
        }

        async fn exists(&self, _courier_id: Uuid) -> Result<bool, CacheError> {
            Ok(true)
        }

        async fn update_status(&self, _courier_id: Uuid, _status: CourierStatus) -> Result<(), CacheError> {
            Ok(())
        }

        async fn update_max_load(&self, _courier_id: Uuid, _max_load: u32) -> Result<(), CacheError> {
            Ok(())
        }

        async fn add_to_free_pool(&self, _courier_id: Uuid, _work_zone: &str) -> Result<(), CacheError> {
            Ok(())
        }

        async fn remove_from_free_pool(&self, _courier_id: Uuid, _work_zone: &str) -> Result<(), CacheError> {
            Ok(())
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
            _filter: PackageFilter,
            _limit: u64,
            _offset: u64,
        ) -> Result<Vec<Package>, RepositoryError> {
            Ok(vec![])
        }

        async fn count_by_filter(&self, _filter: PackageFilter) -> Result<u64, RepositoryError> {
            Ok(0)
        }

        async fn find_by_courier(&self, _courier_id: Uuid) -> Result<Vec<Package>, RepositoryError> {
            Ok(vec![])
        }

        async fn delete(&self, _id: PackageId) -> Result<(), RepositoryError> {
            Ok(())
        }
    }

    // ==================== Mock Notification Service ====================

    struct MockNotificationService;

    #[async_trait]
    impl NotificationService for MockNotificationService {
        async fn send_order_assigned(&self, _push_token: &str, _notification: OrderAssignedNotification) -> Result<(), NotificationError> {
            Ok(())
        }
        async fn send_delivery_status(&self, _push_token: &str, _notification: DeliveryStatusNotification) -> Result<(), NotificationError> {
            Ok(())
        }
    }

    // ==================== Mock Location Cache ====================

    struct MockLocationCache;

    #[async_trait]
    impl LocationCache for MockLocationCache {
        async fn set_location(&self, _location: &CourierLocation) -> Result<(), LocationCacheError> {
            Ok(())
        }
        async fn get_location(&self, _courier_id: Uuid) -> Result<Option<CourierLocation>, LocationCacheError> {
            Ok(None)
        }
        async fn get_locations(&self, _courier_ids: &[Uuid]) -> Result<Vec<CourierLocation>, LocationCacheError> {
            Ok(vec![])
        }
        async fn delete_location(&self, _courier_id: Uuid) -> Result<(), LocationCacheError> {
            Ok(())
        }
        async fn has_location(&self, _courier_id: Uuid) -> Result<bool, LocationCacheError> {
            Ok(false)
        }
        async fn get_all_locations(&self) -> Result<Vec<CourierLocation>, LocationCacheError> {
            Ok(vec![])
        }
        async fn get_active_courier_ids(&self) -> Result<Vec<Uuid>, LocationCacheError> {
            Ok(vec![])
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

    fn create_test_package_in_pool() -> Package {
        let now = Utc::now();
        let period = DeliveryPeriod::new(
            now + chrono::Duration::hours(2),
            now + chrono::Duration::hours(4),
        )
        .unwrap();

        let mut package = Package::new(
            Uuid::new_v4(),
            Uuid::new_v4(),
            Some("+49123456789".to_string()), // customer_phone
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

        // Move to pool (required for assignment)
        package.move_to_pool().unwrap();
        package
    }

    fn create_test_courier(id: Uuid) -> Courier {
        use crate::domain::model::courier::WorkHours;
        use chrono::NaiveTime;

        let work_hours = WorkHours::new(
            NaiveTime::from_hms_opt(9, 0, 0).unwrap(),
            NaiveTime::from_hms_opt(18, 0, 0).unwrap(),
            vec![1, 2, 3, 4, 5],
        )
        .unwrap();

        let mut courier = Courier::builder(
            "Test Courier".to_string(),
            "+1234567890".to_string(),
            "test@example.com".to_string(),
            TransportType::Car,
            50.0,
            "Berlin-101".to_string(),
            work_hours,
        )
        .build()
        .unwrap();

        // Override the ID using reconstitute pattern
        // For testing, we need to set a specific ID
        Courier::reconstitute(
            CourierId::from_uuid(id),
            courier.name().to_string(),
            courier.phone().to_string(),
            courier.email().to_string(),
            courier.transport_type(),
            courier.max_distance_km(),
            courier.work_zone().to_string(),
            courier.work_hours().clone(),
            None, // push_token
            courier.status(),
            courier.capacity().clone(),
            courier.rating(),
            courier.successful_deliveries(),
            courier.failed_deliveries(),
            courier.created_at(),
            courier.updated_at(),
            courier.version(),
        )
    }

    // ==================== Tests ====================

    fn create_handler(
        courier_repo: Arc<MockCourierRepository>,
        courier_cache: Arc<MockCourierCache>,
        package_repo: Arc<MockPackageRepository>,
    ) -> Handler<MockCourierRepository, MockCourierCache, MockPackageRepository, MockEventPublisher, MockNotificationService, MockLocationCache> {
        Handler::new(
            courier_repo,
            courier_cache,
            package_repo,
            Arc::new(MockEventPublisher),
            Arc::new(MockNotificationService),
            Arc::new(MockLocationCache),
        )
    }

    #[tokio::test]
    async fn test_assign_order_manual_success() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::new());

        // Setup: Create package and courier
        let package = create_test_package_in_pool();
        let package_id = package.id().0;
        package_repo.add_package(package);

        let courier_id = Uuid::new_v4();
        let mut courier = create_test_courier(courier_id);
        courier.go_online().unwrap(); // Set to FREE status
        courier_repo.add_courier(courier);

        let handler = create_handler(courier_repo, courier_cache, package_repo.clone());

        let cmd = Command::manual_assign(package_id, courier_id);
        let result = handler.handle(cmd).await;

        assert!(result.is_ok(), "Expected Ok, got: {:?}", result.err());
        let response = result.unwrap();
        assert_eq!(response.package_id, package_id);
        assert_eq!(response.courier_id, courier_id);

        // Verify package status changed
        let updated_package = package_repo.find_by_id(PackageId::from_uuid(package_id)).await.unwrap();
        assert!(updated_package.is_some());
        assert_eq!(updated_package.unwrap().status(), PackageStatus::Assigned);
    }

    #[tokio::test]
    async fn test_assign_order_package_not_found() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::new());

        let handler = create_handler(courier_repo, courier_cache, package_repo);

        let cmd = Command::auto_assign(Uuid::new_v4());
        let result = handler.handle(cmd).await;

        assert!(matches!(result, Err(AssignOrderError::PackageNotFound(_))));
    }

    #[tokio::test]
    async fn test_assign_order_package_wrong_status() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::new());

        // Create package but don't move to pool (status is ACCEPTED)
        let now = Utc::now();
        let period = DeliveryPeriod::new(
            now + chrono::Duration::hours(2),
            now + chrono::Duration::hours(4),
        )
        .unwrap();

        let package = Package::new(
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
        let package_id = package.id().0;
        package_repo.add_package(package);

        let handler = create_handler(courier_repo, courier_cache, package_repo);

        let cmd = Command::auto_assign(package_id);
        let result = handler.handle(cmd).await;

        assert!(matches!(result, Err(AssignOrderError::InvalidPackageStatus(_))));
    }

    #[tokio::test]
    async fn test_assign_order_no_couriers() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::new());

        let package = create_test_package_in_pool();
        let package_id = package.id().0;
        package_repo.add_package(package);

        // No couriers in cache
        let handler = create_handler(courier_repo, courier_cache, package_repo);

        let cmd = Command::auto_assign(package_id);
        let result = handler.handle(cmd).await;

        assert!(matches!(result, Err(AssignOrderError::NoAvailableCourier(_))));
    }

    #[tokio::test]
    async fn test_assign_order_courier_not_found() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache::new());
        let package_repo = Arc::new(MockPackageRepository::new());

        let package = create_test_package_in_pool();
        let package_id = package.id().0;
        package_repo.add_package(package);

        let handler = create_handler(courier_repo, courier_cache, package_repo);

        // Manual assign with non-existent courier
        let cmd = Command::manual_assign(package_id, Uuid::new_v4());
        let result = handler.handle(cmd).await;

        assert!(matches!(result, Err(AssignOrderError::CourierNotFound(_))));
    }
}
