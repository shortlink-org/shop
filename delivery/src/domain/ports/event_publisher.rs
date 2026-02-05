//! Event Publisher Port
//!
//! Defines the interface for publishing domain events to a message broker.
//! This port is implemented by infrastructure adapters (e.g., Kafka).
//! Topic mapping and transport details belong in the infrastructure layer.

use async_trait::async_trait;
use thiserror::Error;

use crate::domain::model::domain::delivery::events::v1::{
    CourierLocationUpdatedEvent, CourierRegisteredEvent, CourierStatusChangedEvent,
    PackageAcceptedEvent, PackageAssignedEvent, PackageDeliveredEvent, PackageInTransitEvent,
    PackageNotDeliveredEvent, PackageRequiresHandlingEvent,
};

/// Domain-agnostic errors that can occur during event publishing.
/// Transport-specific failures (connection, timeout, etc.) are mapped to these by adapters.
#[derive(Debug, Error)]
pub enum EventPublisherError {
    /// Event could not be published (e.g. broker unavailable, timeout, send failure).
    #[error("Publish failed: {0}")]
    PublishFailed(String),

    /// Event payload could not be serialized.
    #[error("Serialization failed: {0}")]
    SerializationFailed(String),
}

/// Sum type of all domain events that can be published.
/// Topic routing is decided by the infrastructure adapter, not the domain.
#[derive(Debug, Clone)]
pub enum DomainEvent {
    PackageAccepted(PackageAcceptedEvent),
    PackageAssigned(PackageAssignedEvent),
    PackageInTransit(PackageInTransitEvent),
    PackageDelivered(PackageDeliveredEvent),
    PackageNotDelivered(PackageNotDeliveredEvent),
    PackageRequiresHandling(PackageRequiresHandlingEvent),
    CourierRegistered(CourierRegisteredEvent),
    CourierLocationUpdated(CourierLocationUpdatedEvent),
    CourierStatusChanged(CourierStatusChangedEvent),
}

/// Event Publisher Port
///
/// Defines the contract for publishing domain events.
/// Implementations handle transport and topic mapping (Kafka, NATS, Outbox, etc.).
#[async_trait]
pub trait EventPublisher: Send + Sync {
    /// Publish a domain event. Routing (e.g. topic selection) is the responsibility of the adapter.
    async fn publish(&self, event: DomainEvent) -> Result<(), EventPublisherError>;
}

#[cfg(test)]
/// Shared mock for tests. Implements EventPublisher with a no-op publish.
pub struct MockEventPublisher;

#[cfg(test)]
#[async_trait]
impl EventPublisher for MockEventPublisher {
    async fn publish(&self, _event: DomainEvent) -> Result<(), EventPublisherError> {
        Ok(())
    }
}
