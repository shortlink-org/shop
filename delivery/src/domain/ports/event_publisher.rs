//! Event Publisher Port
//!
//! Defines the interface for publishing domain events to a message broker.
//! This port is implemented by infrastructure adapters (e.g., Kafka).

use async_trait::async_trait;
#[cfg(test)]
use mockall::automock;
use thiserror::Error;

use crate::domain::model::domain::delivery::events::v1::{
    CourierLocationUpdatedEvent, CourierRegisteredEvent, CourierStatusChangedEvent,
    PackageAcceptedEvent, PackageAssignedEvent, PackageDeliveredEvent, PackageInTransitEvent,
    PackageNotDeliveredEvent, PackageRequiresHandlingEvent,
};

/// Errors that can occur during event publishing
#[derive(Debug, Error)]
pub enum EventPublisherError {
    /// Connection error to message broker
    #[error("Connection error: {0}")]
    ConnectionError(String),

    /// Serialization error
    #[error("Serialization error: {0}")]
    SerializationError(String),

    /// Publish error
    #[error("Failed to publish event: {0}")]
    PublishError(String),

    /// Timeout error
    #[error("Publish timeout: {0}")]
    Timeout(String),
}

/// Event Publisher Port
///
/// Defines the contract for publishing domain events.
/// Implementations handle the actual message broker integration (Kafka, etc.).
#[cfg_attr(test, automock)]
#[async_trait]
pub trait EventPublisher: Send + Sync {
    /// Publish a PackageAccepted event
    ///
    /// Topic: delivery.package.accepted.v1
    async fn publish_package_accepted(
        &self,
        event: PackageAcceptedEvent,
    ) -> Result<(), EventPublisherError>;

    /// Publish a PackageAssigned event
    ///
    /// Topic: delivery.order.assigned.v1
    async fn publish_package_assigned(
        &self,
        event: PackageAssignedEvent,
    ) -> Result<(), EventPublisherError>;

    /// Publish a PackageInTransit event
    ///
    /// Topic: delivery.package.in_transit.v1
    async fn publish_package_in_transit(
        &self,
        event: PackageInTransitEvent,
    ) -> Result<(), EventPublisherError>;

    /// Publish a PackageDelivered event
    ///
    /// Topic: delivery.package.delivered.v1
    async fn publish_package_delivered(
        &self,
        event: PackageDeliveredEvent,
    ) -> Result<(), EventPublisherError>;

    /// Publish a PackageNotDelivered event
    ///
    /// Topic: delivery.package.not_delivered.v1
    async fn publish_package_not_delivered(
        &self,
        event: PackageNotDeliveredEvent,
    ) -> Result<(), EventPublisherError>;

    /// Publish a PackageRequiresHandling event
    ///
    /// Topic: delivery.package.requires_handling.v1
    async fn publish_package_requires_handling(
        &self,
        event: PackageRequiresHandlingEvent,
    ) -> Result<(), EventPublisherError>;

    /// Publish a CourierRegistered event
    ///
    /// Topic: delivery.courier.registered.v1
    async fn publish_courier_registered(
        &self,
        event: CourierRegisteredEvent,
    ) -> Result<(), EventPublisherError>;

    /// Publish a CourierLocationUpdated event
    ///
    /// Topic: delivery.courier.location_updated.v1
    async fn publish_courier_location_updated(
        &self,
        event: CourierLocationUpdatedEvent,
    ) -> Result<(), EventPublisherError>;

    /// Publish a CourierStatusChanged event
    ///
    /// Topic: delivery.courier.status_changed.v1
    async fn publish_courier_status_changed(
        &self,
        event: CourierStatusChangedEvent,
    ) -> Result<(), EventPublisherError>;
}
