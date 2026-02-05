//! Pick Up Order Handler
//!
//! Handles confirming package pickup by courier.
//!
//! ## Flow
//! 1. Load package from repository
//! 2. Validate package status is ASSIGNED
//! 3. Validate courier is assigned to package
//! 4. Start transit (ASSIGNED -> IN_TRANSIT)
//! 5. Save package
//! 6. Publish PackageInTransitEvent
//! 7. Return response

use std::sync::Arc;

use chrono::{DateTime, Utc};
use thiserror::Error;
use tracing::{info, warn};
use uuid::Uuid;

use crate::domain::model::domain::delivery::common::v1 as proto_common;
use crate::domain::model::domain::delivery::events::v1::PackageInTransitEvent;
use crate::domain::model::package::{PackageId, PackageStatus};
use crate::domain::ports::{
    CommandHandlerWithResult, EventPublisher, GeolocationService, PackageRepository,
    RepositoryError,
};

use super::Command;

/// Errors that can occur during package pickup
#[derive(Debug, Error)]
pub enum PickUpOrderError {
    /// Package not found
    #[error("Package not found: {0}")]
    PackageNotFound(Uuid),

    /// Courier not assigned to package
    #[error("Courier {0} is not assigned to package {1}")]
    CourierNotAssigned(Uuid, Uuid),

    /// Invalid package status
    #[error("Invalid package status for pickup: expected Assigned, got {0}")]
    InvalidPackageStatus(PackageStatus),

    /// Package already picked up
    #[error("Package already picked up (in transit): {0}")]
    AlreadyPickedUp(Uuid),

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),
}

/// Response from picking up an order
#[derive(Debug, Clone)]
pub struct Response {
    /// The package ID
    pub package_id: Uuid,
    /// The updated package status
    pub status: PackageStatus,
    /// When the package was picked up
    pub picked_up_at: DateTime<Utc>,
}

/// Pick Up Order Handler
pub struct Handler<P, E, G>
where
    P: PackageRepository,
    E: EventPublisher,
    G: GeolocationService,
{
    package_repo: Arc<P>,
    event_publisher: Arc<E>,
    geolocation_service: Arc<G>,
}

impl<P, E, G> Handler<P, E, G>
where
    P: PackageRepository,
    E: EventPublisher,
    G: GeolocationService,
{
    /// Create a new handler instance
    pub fn new(
        package_repo: Arc<P>,
        event_publisher: Arc<E>,
        geolocation_service: Arc<G>,
    ) -> Self {
        Self {
            package_repo,
            event_publisher,
            geolocation_service,
        }
    }
}

impl<P, E, G> CommandHandlerWithResult<Command, Response> for Handler<P, E, G>
where
    P: PackageRepository + Send + Sync,
    E: EventPublisher + Send + Sync,
    G: GeolocationService + Send + Sync,
{
    type Error = PickUpOrderError;

    /// Handle the PickUpOrder command
    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        // 1. Load package from repository
        let mut package = self
            .package_repo
            .find_by_id(PackageId::from_uuid(cmd.package_id))
            .await?
            .ok_or(PickUpOrderError::PackageNotFound(cmd.package_id))?;

        // 2. Validate package status
        match package.status() {
            PackageStatus::Assigned => {}
            PackageStatus::InTransit => {
                return Err(PickUpOrderError::AlreadyPickedUp(cmd.package_id));
            }
            status => {
                return Err(PickUpOrderError::InvalidPackageStatus(status));
            }
        }

        // 3. Validate courier is assigned to this package
        let assigned_courier = package
            .courier_id()
            .ok_or(PickUpOrderError::CourierNotAssigned(
                cmd.courier_id,
                cmd.package_id,
            ))?;

        if assigned_courier != cmd.courier_id {
            return Err(PickUpOrderError::CourierNotAssigned(
                cmd.courier_id,
                cmd.package_id,
            ));
        }

        // 4. Start transit
        package.start_transit().map_err(|_| {
            PickUpOrderError::InvalidPackageStatus(package.status())
        })?;

        let picked_up_at = package.updated_at();

        // 5. Save package to repository
        self.package_repo.save(&package).await?;

        // 6. Publish PackageInTransitEvent
        let now = Utc::now();
        let event = PackageInTransitEvent {
            package_id: package.id().0.to_string(),
            order_id: package.order_id().to_string(),
            courier_id: cmd.courier_id.to_string(),
            status: proto_common::PackageStatus::InTransit as i32,
            in_transit_at: Some(pbjson_types::Timestamp {
                seconds: now.timestamp(),
                nanos: now.timestamp_subsec_nanos() as i32,
            }),
            courier_location: Some(proto_common::Location {
                latitude: cmd.pickup_location.latitude(),
                longitude: cmd.pickup_location.longitude(),
                accuracy: cmd.pickup_location.accuracy(),
                timestamp: None,
                speed: None,
                heading: None,
            }),
            occurred_at: Some(pbjson_types::Timestamp {
                seconds: now.timestamp(),
                nanos: now.timestamp_subsec_nanos() as i32,
            }),
        };

        if let Err(e) = self.event_publisher.publish_package_in_transit(event).await {
            warn!(
                package_id = %cmd.package_id,
                courier_id = %cmd.courier_id,
                error = %e,
                "Failed to publish PackageInTransit event (non-fatal)"
            );
        } else {
            info!(
                package_id = %cmd.package_id,
                courier_id = %cmd.courier_id,
                "PackageInTransit event published"
            );
        }

        // 7. Update courier location via GeolocationService
        if let Err(e) = self
            .geolocation_service
            .update_location(cmd.courier_id, cmd.pickup_location, now)
            .await
        {
            warn!(
                package_id = %cmd.package_id,
                courier_id = %cmd.courier_id,
                error = %e,
                "Failed to update courier location via GeolocationService (non-fatal)"
            );
        }

        Ok(Response {
            package_id: cmd.package_id,
            status: package.status(),
            picked_up_at,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::domain::delivery::events::v1::{
        CourierLocationUpdatedEvent, CourierRegisteredEvent, CourierStatusChangedEvent,
        PackageAcceptedEvent, PackageAssignedEvent, PackageDeliveredEvent,
        PackageNotDeliveredEvent, PackageRequiresHandlingEvent,
    };
    use crate::domain::model::package::{Address, DeliveryPeriod, Package, Priority};
    use crate::domain::model::vo::location::Location;
    use crate::domain::ports::{
        EventPublisherError, GeolocationService, GeolocationServiceError, PackageFilter,
    };
    use async_trait::async_trait;
    use chrono::Utc;
    use std::collections::HashMap;
    use std::sync::Mutex;

    // ==================== Mock Event Publisher ====================

    struct MockEventPublisher;

    #[async_trait]
    impl EventPublisher for MockEventPublisher {
        async fn publish_package_accepted(&self, _event: PackageAcceptedEvent) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_package_assigned(&self, _event: PackageAssignedEvent) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_package_in_transit(&self, _event: PackageInTransitEvent) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_package_delivered(&self, _event: PackageDeliveredEvent) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_package_not_delivered(&self, _event: PackageNotDeliveredEvent) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_package_requires_handling(&self, _event: PackageRequiresHandlingEvent) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_courier_registered(&self, _event: CourierRegisteredEvent) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_courier_location_updated(&self, _event: CourierLocationUpdatedEvent) -> Result<(), EventPublisherError> {
            Ok(())
        }
        async fn publish_courier_status_changed(&self, _event: CourierStatusChangedEvent) -> Result<(), EventPublisherError> {
            Ok(())
        }
    }

    struct MockGeolocationService;

    #[async_trait]
    impl GeolocationService for MockGeolocationService {
        async fn update_location(
            &self,
            _courier_id: Uuid,
            _location: Location,
            _timestamp: chrono::DateTime<Utc>,
        ) -> Result<(), GeolocationServiceError> {
            Ok(())
        }

        async fn get_location(
            &self,
            _courier_id: Uuid,
        ) -> Result<Option<crate::domain::model::CourierLocation>, GeolocationServiceError> {
            Ok(None)
        }
    }

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

        // Move to pool -> assign
        package.move_to_pool().unwrap();
        package.assign_to(courier_id).unwrap();
        package
    }

    fn create_in_transit_package(courier_id: Uuid) -> Package {
        let mut package = create_assigned_package(courier_id);
        package.start_transit().unwrap();
        package
    }

    // ==================== Tests ====================

    #[tokio::test]
    async fn test_pick_up_order_success() {
        let package_repo = Arc::new(MockPackageRepository::new());
        let event_publisher = Arc::new(MockEventPublisher);

        let courier_id = Uuid::new_v4();
        let package = create_assigned_package(courier_id);
        let package_id = package.id().0;
        package_repo.add_package(package);

        let handler = Handler::new(
            package_repo.clone(),
            event_publisher,
            Arc::new(MockGeolocationService),
        );

        let cmd = Command::new(package_id, courier_id, create_test_location());
        let result = handler.handle(cmd).await;

        assert!(result.is_ok(), "Expected Ok, got: {:?}", result.err());
        let response = result.unwrap();
        assert_eq!(response.package_id, package_id);
        assert_eq!(response.status, PackageStatus::InTransit);

        // Verify package status changed
        let updated_package = package_repo
            .find_by_id(PackageId::from_uuid(package_id))
            .await
            .unwrap();
        assert!(updated_package.is_some());
        assert_eq!(updated_package.unwrap().status(), PackageStatus::InTransit);
    }

    #[tokio::test]
    async fn test_pick_up_order_package_not_found() {
        let package_repo = Arc::new(MockPackageRepository::new());
        let event_publisher = Arc::new(MockEventPublisher);
        let handler = Handler::new(
            package_repo,
            event_publisher,
            Arc::new(MockGeolocationService),
        );

        let cmd = Command::new(Uuid::new_v4(), Uuid::new_v4(), create_test_location());
        let result = handler.handle(cmd).await;

        assert!(matches!(result, Err(PickUpOrderError::PackageNotFound(_))));
    }

    #[tokio::test]
    async fn test_pick_up_order_already_picked_up() {
        let package_repo = Arc::new(MockPackageRepository::new());
        let event_publisher = Arc::new(MockEventPublisher);

        let courier_id = Uuid::new_v4();
        let package = create_in_transit_package(courier_id);
        let package_id = package.id().0;
        package_repo.add_package(package);

        let handler = Handler::new(
            package_repo,
            event_publisher,
            Arc::new(MockGeolocationService),
        );

        let cmd = Command::new(package_id, courier_id, create_test_location());
        let result = handler.handle(cmd).await;

        assert!(matches!(result, Err(PickUpOrderError::AlreadyPickedUp(_))));
    }

    #[tokio::test]
    async fn test_pick_up_order_courier_not_assigned() {
        let package_repo = Arc::new(MockPackageRepository::new());
        let event_publisher = Arc::new(MockEventPublisher);

        let assigned_courier_id = Uuid::new_v4();
        let wrong_courier_id = Uuid::new_v4();

        let package = create_assigned_package(assigned_courier_id);
        let package_id = package.id().0;
        package_repo.add_package(package);

        let handler = Handler::new(
            package_repo,
            event_publisher,
            Arc::new(MockGeolocationService),
        );

        // Try to pick up with wrong courier
        let cmd = Command::new(package_id, wrong_courier_id, create_test_location());
        let result = handler.handle(cmd).await;

        assert!(matches!(
            result,
            Err(PickUpOrderError::CourierNotAssigned(_, _))
        ));
    }

    #[tokio::test]
    async fn test_pick_up_order_invalid_status() {
        let package_repo = Arc::new(MockPackageRepository::new());
        let event_publisher = Arc::new(MockEventPublisher);

        let courier_id = Uuid::new_v4();

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
            package_repo,
            event_publisher,
            Arc::new(MockGeolocationService),
        );

        let cmd = Command::new(package_id, courier_id, create_test_location());
        let result = handler.handle(cmd).await;

        assert!(matches!(
            result,
            Err(PickUpOrderError::InvalidPackageStatus(_))
        ));
    }
}
