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

use crate::domain::model::domain::delivery::common::v1 as proto_common;
use crate::domain::model::domain::delivery::events::v1::{
    PackageDeliveredEvent, PackageNotDeliveredEvent,
};
use crate::domain::model::package::{PackageId, PackageStatus};
use crate::domain::ports::{
    CommandHandlerWithResult, CourierCache, CourierRepository, EventPublisher, PackageRepository,
    RepositoryError,
};

use super::command::{DeliveryResult, NotDeliveredReason};
use super::Command;

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

    /// Missing reason for failed delivery
    #[error("Reason required for failed delivery")]
    MissingNotDeliveredReason,

    /// Package already delivered
    #[error("Package already delivered: {0}")]
    AlreadyDelivered(Uuid),

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
pub struct Handler<R, C, P, E>
where
    R: CourierRepository,
    C: CourierCache,
    P: PackageRepository,
    E: EventPublisher,
{
    courier_repo: Arc<R>,
    courier_cache: Arc<C>,
    package_repo: Arc<P>,
    event_publisher: Arc<E>,
}

impl<R, C, P, E> Handler<R, C, P, E>
where
    R: CourierRepository,
    C: CourierCache,
    P: PackageRepository,
    E: EventPublisher,
{
    /// Create a new handler instance
    pub fn new(
        courier_repo: Arc<R>,
        courier_cache: Arc<C>,
        package_repo: Arc<P>,
        event_publisher: Arc<E>,
    ) -> Self {
        Self {
            courier_repo,
            courier_cache,
            package_repo,
            event_publisher,
        }
    }

    /// Convert NotDeliveredReason to string for storage
    fn reason_to_string(reason: &NotDeliveredReason) -> String {
        match reason {
            NotDeliveredReason::CustomerUnavailable => "CUSTOMER_NOT_AVAILABLE".to_string(),
            NotDeliveredReason::WrongAddress => "WRONG_ADDRESS".to_string(),
            NotDeliveredReason::Refused => "CUSTOMER_REFUSED".to_string(),
            NotDeliveredReason::AccessDenied => "ACCESS_DENIED".to_string(),
            NotDeliveredReason::Other(desc) => format!("OTHER: {}", desc),
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
}

impl<R, C, P, E> CommandHandlerWithResult<Command, Response> for Handler<R, C, P, E>
where
    R: CourierRepository + Send + Sync,
    C: CourierCache + Send + Sync,
    P: PackageRepository + Send + Sync,
    E: EventPublisher + Send + Sync,
{
    type Error = DeliverOrderError;

    /// Handle the DeliverOrder command
    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        // 1. Validate reason is provided if result is NotDelivered
        if cmd.result == DeliveryResult::NotDelivered && cmd.not_delivered_reason.is_none() {
            return Err(DeliverOrderError::MissingNotDeliveredReason);
        }

        // 2. Load package from repository
        let mut package = self
            .package_repo
            .find_by_id(PackageId::from_uuid(cmd.package_id))
            .await?
            .ok_or(DeliverOrderError::PackageNotFound(cmd.package_id))?;

        // 3. Validate package status (must be Assigned or InTransit)
        match package.status() {
            PackageStatus::Assigned | PackageStatus::InTransit => {}
            PackageStatus::Delivered => {
                return Err(DeliverOrderError::AlreadyDelivered(cmd.package_id));
            }
            status => {
                return Err(DeliverOrderError::InvalidPackageStatus(status));
            }
        }

        // 4. Validate courier is assigned to this package
        let assigned_courier = package
            .courier_id()
            .ok_or(DeliverOrderError::CourierNotAssigned(
                cmd.courier_id,
                cmd.package_id,
            ))?;

        if assigned_courier != cmd.courier_id {
            return Err(DeliverOrderError::CourierNotAssigned(
                cmd.courier_id,
                cmd.package_id,
            ));
        }

        // 5. Load courier to get current stats
        let courier = self
            .courier_repo
            .find_by_id(cmd.courier_id)
            .await?
            .ok_or(DeliverOrderError::CourierNotFound(cmd.courier_id))?;

        // 6. Update package status and courier stats based on result
        match cmd.result {
            DeliveryResult::Delivered => {
                // Mark package as delivered
                package.mark_delivered().map_err(|e| {
                    DeliverOrderError::InvalidPackageStatus(package.status())
                })?;

                // Update courier stats - increment successful deliveries
                self.courier_cache
                    .update_stats(
                        cmd.courier_id,
                        courier.rating(),
                        courier.successful_deliveries() + 1,
                        courier.failed_deliveries(),
                    )
                    .await
                    .map_err(|e| {
                        RepositoryError::QueryError(format!("Failed to update courier stats: {}", e))
                    })?;
            }
            DeliveryResult::NotDelivered => {
                // Mark package as not delivered with reason
                let reason = cmd
                    .not_delivered_reason
                    .as_ref()
                    .map(Self::reason_to_string)
                    .unwrap_or_default();

                package.mark_not_delivered(reason).map_err(|e| {
                    DeliverOrderError::InvalidPackageStatus(package.status())
                })?;

                // Update courier stats - increment failed deliveries
                self.courier_cache
                    .update_stats(
                        cmd.courier_id,
                        courier.rating(),
                        courier.successful_deliveries(),
                        courier.failed_deliveries() + 1,
                    )
                    .await
                    .map_err(|e| {
                        RepositoryError::QueryError(format!("Failed to update courier stats: {}", e))
                    })?;
            }
        }

        // 7. Update courier load (decrement by 1)
        let new_load = courier.current_load().saturating_sub(1);
        self.courier_cache
            .update_load(cmd.courier_id, new_load, courier.max_load())
            .await
            .map_err(|e| {
                RepositoryError::QueryError(format!("Failed to update courier load: {}", e))
            })?;

        // 8. Save package to repository
        self.package_repo.save(&package).await?;

        // 9. Publish delivery event
        let now = Utc::now();
        match cmd.result {
            DeliveryResult::Delivered => {
                let event = PackageDeliveredEvent {
                    package_id: package.id().0.to_string(),
                    order_id: package.order_id().to_string(),
                    courier_id: cmd.courier_id.to_string(),
                    status: proto_common::PackageStatus::Delivered as i32,
                    delivered_at: Some(pbjson_types::Timestamp {
                        seconds: now.timestamp(),
                        nanos: now.timestamp_subsec_nanos() as i32,
                    }),
                    delivery_location: Some(proto_common::Location {
                        latitude: cmd.confirmation_location.latitude(),
                        longitude: cmd.confirmation_location.longitude(),
                        accuracy: cmd.confirmation_location.accuracy(),
                        timestamp: None,
                        speed: None,
                        heading: None,
                    }),
                    photo: cmd.photo_proof.unwrap_or_default(),
                    customer_signature: cmd.signature.unwrap_or_default(),
                    occurred_at: Some(pbjson_types::Timestamp {
                        seconds: now.timestamp(),
                        nanos: now.timestamp_subsec_nanos() as i32,
                    }),
                };

                if let Err(e) = self.event_publisher.publish_package_delivered(event).await {
                    warn!(
                        package_id = %cmd.package_id,
                        courier_id = %cmd.courier_id,
                        error = %e,
                        "Failed to publish PackageDelivered event (non-fatal)"
                    );
                } else {
                    info!(
                        package_id = %cmd.package_id,
                        courier_id = %cmd.courier_id,
                        "PackageDelivered event published"
                    );
                }
            }
            DeliveryResult::NotDelivered => {
                let reason = cmd
                    .not_delivered_reason
                    .as_ref()
                    .map(Self::reason_to_proto)
                    .unwrap_or(proto_common::NotDeliveredReason::Unspecified);

                let reason_description = cmd
                    .not_delivered_reason
                    .as_ref()
                    .map(Self::get_reason_description)
                    .unwrap_or_default();

                let event = PackageNotDeliveredEvent {
                    package_id: package.id().0.to_string(),
                    order_id: package.order_id().to_string(),
                    courier_id: cmd.courier_id.to_string(),
                    status: proto_common::PackageStatus::NotDelivered as i32,
                    reason: reason as i32,
                    reason_description,
                    not_delivered_at: Some(pbjson_types::Timestamp {
                        seconds: now.timestamp(),
                        nanos: now.timestamp_subsec_nanos() as i32,
                    }),
                    courier_location: Some(proto_common::Location {
                        latitude: cmd.confirmation_location.latitude(),
                        longitude: cmd.confirmation_location.longitude(),
                        accuracy: cmd.confirmation_location.accuracy(),
                        timestamp: None,
                        speed: None,
                        heading: None,
                    }),
                    occurred_at: Some(pbjson_types::Timestamp {
                        seconds: now.timestamp(),
                        nanos: now.timestamp_subsec_nanos() as i32,
                    }),
                };

                if let Err(e) = self
                    .event_publisher
                    .publish_package_not_delivered(event)
                    .await
                {
                    warn!(
                        package_id = %cmd.package_id,
                        courier_id = %cmd.courier_id,
                        reason = ?cmd.not_delivered_reason,
                        error = %e,
                        "Failed to publish PackageNotDelivered event (non-fatal)"
                    );
                } else {
                    info!(
                        package_id = %cmd.package_id,
                        courier_id = %cmd.courier_id,
                        reason = ?cmd.not_delivered_reason,
                        "PackageNotDelivered event published"
                    );
                }
            }
        }

        // 10. TODO: Notify OMS about delivery result
        // 11. TODO: Update courier location via Geolocation service

        Ok(Response {
            package_id: cmd.package_id,
            status: package.status(),
            updated_at: package.updated_at(),
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::courier::{Courier, CourierId, WorkHours};
    use crate::domain::model::domain::delivery::events::v1::{
        CourierLocationUpdatedEvent, CourierRegisteredEvent, CourierStatusChangedEvent,
        PackageAcceptedEvent, PackageAssignedEvent, PackageInTransitEvent,
        PackageRequiresHandlingEvent,
    };
    use crate::domain::model::package::{Address, DeliveryPeriod, Package, Priority};
    use crate::domain::model::vo::location::Location;
    use crate::domain::model::vo::TransportType;
    use crate::domain::ports::{CacheError, CachedCourierState, EventPublisherError, PackageFilter};
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
    }

    struct MockCourierCache;

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
            _status: crate::domain::model::courier::CourierStatus,
            _work_zone: &str,
        ) -> Result<(), CacheError> {
            Ok(())
        }

        async fn get_status(
            &self,
            _courier_id: Uuid,
        ) -> Result<Option<crate::domain::model::courier::CourierStatus>, CacheError> {
            Ok(None)
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

        async fn update_status(
            &self,
            _courier_id: Uuid,
            _status: crate::domain::model::courier::CourierStatus,
        ) -> Result<(), CacheError> {
            Ok(())
        }

        async fn update_max_load(&self, _courier_id: Uuid, _max_load: u32) -> Result<(), CacheError> {
            Ok(())
        }

        async fn add_to_free_pool(&self, _courier_id: Uuid, _work_zone: &str) -> Result<(), CacheError> {
            Ok(())
        }

        async fn remove_from_free_pool(
            &self,
            _courier_id: Uuid,
            _work_zone: &str,
        ) -> Result<(), CacheError> {
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

    // ==================== Mock Event Publisher ====================

    struct MockEventPublisher;

    #[async_trait]
    impl EventPublisher for MockEventPublisher {
        async fn publish_package_accepted(
            &self,
            _event: PackageAcceptedEvent,
        ) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_package_assigned(
            &self,
            _event: PackageAssignedEvent,
        ) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_package_in_transit(
            &self,
            _event: PackageInTransitEvent,
        ) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_package_delivered(
            &self,
            _event: PackageDeliveredEvent,
        ) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_package_not_delivered(
            &self,
            _event: PackageNotDeliveredEvent,
        ) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_package_requires_handling(
            &self,
            _event: PackageRequiresHandlingEvent,
        ) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_courier_registered(
            &self,
            _event: CourierRegisteredEvent,
        ) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_courier_location_updated(
            &self,
            _event: CourierLocationUpdatedEvent,
        ) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_courier_status_changed(
            &self,
            _event: CourierStatusChangedEvent,
        ) -> Result<(), EventPublisherError> {
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
            crate::domain::model::courier::CourierCapacity::new(5),
            4.5,
            10, // successful_deliveries
            1,  // failed_deliveries
            Utc::now(),
            Utc::now(),
            1,
        )
    }

    // ==================== Tests ====================

    #[tokio::test]
    async fn test_deliver_order_success() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache);
        let package_repo = Arc::new(MockPackageRepository::new());

        let courier_id = Uuid::new_v4();
        let courier = create_test_courier(courier_id);
        courier_repo.add_courier(courier);

        let package = create_in_transit_package(courier_id);
        let package_id = package.id().0;
        package_repo.add_package(package);

        let event_publisher = Arc::new(MockEventPublisher);
        let handler = Handler::new(courier_repo, courier_cache, package_repo.clone(), event_publisher);

        let cmd = Command::delivered(
            package_id,
            courier_id,
            create_test_location(),
            None,
            None,
        );
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
        let courier_cache = Arc::new(MockCourierCache);
        let package_repo = Arc::new(MockPackageRepository::new());

        let courier_id = Uuid::new_v4();
        let courier = create_test_courier(courier_id);
        courier_repo.add_courier(courier);

        let package = create_in_transit_package(courier_id);
        let package_id = package.id().0;
        package_repo.add_package(package);

        let event_publisher = Arc::new(MockEventPublisher);
        let handler = Handler::new(courier_repo, courier_cache, package_repo.clone(), event_publisher);

        let cmd = Command::not_delivered(
            package_id,
            courier_id,
            create_test_location(),
            NotDeliveredReason::CustomerUnavailable,
            None,
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
        let courier_cache = Arc::new(MockCourierCache);
        let package_repo = Arc::new(MockPackageRepository::new());
        let event_publisher = Arc::new(MockEventPublisher);

        let handler = Handler::new(courier_repo, courier_cache, package_repo, event_publisher);

        let cmd = Command::delivered(
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
        let courier_cache = Arc::new(MockCourierCache);
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

        let event_publisher = Arc::new(MockEventPublisher);
        let handler = Handler::new(courier_repo, courier_cache, package_repo, event_publisher);

        // Try to deliver with wrong courier
        let cmd = Command::delivered(
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
    async fn test_deliver_order_missing_reason() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache);
        let package_repo = Arc::new(MockPackageRepository::new());
        let event_publisher = Arc::new(MockEventPublisher);

        let handler = Handler::new(courier_repo, courier_cache, package_repo, event_publisher);

        // Create command with NotDelivered but no reason
        let cmd = Command {
            package_id: Uuid::new_v4(),
            courier_id: Uuid::new_v4(),
            result: DeliveryResult::NotDelivered,
            not_delivered_reason: None, // Missing!
            confirmation_location: create_test_location(),
            photo_proof: None,
            signature: None,
            notes: None,
        };
        let result = handler.handle(cmd).await;

        assert!(matches!(
            result,
            Err(DeliverOrderError::MissingNotDeliveredReason)
        ));
    }

    #[tokio::test]
    async fn test_deliver_order_invalid_status() {
        let courier_repo = Arc::new(MockCourierRepository::new());
        let courier_cache = Arc::new(MockCourierCache);
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

        let event_publisher = Arc::new(MockEventPublisher);
        let handler = Handler::new(courier_repo, courier_cache, package_repo, event_publisher);

        let cmd = Command::delivered(
            package_id,
            courier_id,
            create_test_location(),
            None,
            None,
        );
        let result = handler.handle(cmd).await;

        assert!(matches!(
            result,
            Err(DeliverOrderError::InvalidPackageStatus(_))
        ));
    }
}
