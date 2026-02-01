//! Kafka Event Publisher
//!
//! Implementation of EventPublisher using rdkafka.
//! Publishes domain events to Kafka topics.

use async_trait::async_trait;
use prost::Message;
use rdkafka::producer::{FutureProducer, FutureRecord};
use rdkafka::ClientConfig;
use std::time::Duration;
use tracing::{error, info, instrument};

use crate::domain::model::domain::delivery::events::v1::{
    CourierLocationUpdatedEvent, CourierRegisteredEvent, CourierStatusChangedEvent,
    PackageAcceptedEvent, PackageAssignedEvent, PackageDeliveredEvent, PackageInTransitEvent,
    PackageNotDeliveredEvent, PackageRequiresHandlingEvent,
};
use crate::domain::ports::{EventPublisher, EventPublisherError};

/// Kafka topic names (consolidated)
/// Format: {domain}.{entity}.{event_group}.v1
///
/// Topics are consolidated by entity lifecycle:
/// - package.status: all package status changes (accepted, in_transit, delivered, not_delivered, requires_handling)
/// - order.assigned: package assignment to courier (separate due to business importance)
/// - courier.events: courier registration and status changes
/// - courier.location: high-frequency location updates (separate for scalability)
pub mod topics {
    /// All package status change events
    /// Events: PackageAccepted, PackageInTransit, PackageDelivered, PackageNotDelivered, PackageRequiresHandling
    pub const PACKAGE_STATUS: &str = "delivery.package.status.v1";

    /// Package assigned to courier
    /// Kept separate due to business importance and different consumers
    pub const ORDER_ASSIGNED: &str = "delivery.order.assigned.v1";

    /// Courier lifecycle events (registration, status changes)
    /// Events: CourierRegistered, CourierStatusChanged
    pub const COURIER_EVENTS: &str = "delivery.courier.events.v1";

    /// High-frequency courier location updates
    /// Kept separate for scalability (high volume, different retention)
    pub const COURIER_LOCATION: &str = "delivery.courier.location.v1";
}

/// Configuration for Kafka publisher
#[derive(Debug, Clone)]
pub struct KafkaPublisherConfig {
    /// Kafka bootstrap servers (comma-separated)
    pub brokers: String,
    /// Client ID for this producer
    pub client_id: String,
    /// Message timeout in milliseconds
    pub message_timeout_ms: u64,
    /// Request timeout in milliseconds
    pub request_timeout_ms: u64,
}

impl Default for KafkaPublisherConfig {
    fn default() -> Self {
        Self {
            brokers: "localhost:9092".to_string(),
            client_id: "delivery-service".to_string(),
            message_timeout_ms: 5000,
            request_timeout_ms: 5000,
        }
    }
}

impl KafkaPublisherConfig {
    /// Create config from environment variables
    pub fn from_env() -> Self {
        Self {
            brokers: std::env::var("KAFKA_BROKERS").unwrap_or_else(|_| "localhost:9092".to_string()),
            client_id: std::env::var("KAFKA_CLIENT_ID")
                .unwrap_or_else(|_| "delivery-service".to_string()),
            message_timeout_ms: std::env::var("KAFKA_MESSAGE_TIMEOUT_MS")
                .ok()
                .and_then(|s| s.parse().ok())
                .unwrap_or(5000),
            request_timeout_ms: std::env::var("KAFKA_REQUEST_TIMEOUT_MS")
                .ok()
                .and_then(|s| s.parse().ok())
                .unwrap_or(5000),
        }
    }
}

/// Kafka event publisher implementation
pub struct KafkaEventPublisher {
    producer: FutureProducer,
    timeout: Duration,
}

impl KafkaEventPublisher {
    /// Create a new Kafka event publisher
    pub fn new(config: KafkaPublisherConfig) -> Result<Self, EventPublisherError> {
        let producer: FutureProducer = ClientConfig::new()
            .set("bootstrap.servers", &config.brokers)
            .set("client.id", &config.client_id)
            .set("message.timeout.ms", config.message_timeout_ms.to_string())
            .set("request.timeout.ms", config.request_timeout_ms.to_string())
            .set("acks", "all")
            .set("enable.idempotence", "true")
            .create()
            .map_err(|e| EventPublisherError::ConnectionError(e.to_string()))?;

        info!("Kafka publisher connected to {}", config.brokers);

        Ok(Self {
            producer,
            timeout: Duration::from_millis(config.message_timeout_ms),
        })
    }

    /// Publish a protobuf message to a topic
    async fn publish<M: Message>(
        &self,
        topic: &str,
        key: &str,
        message: &M,
    ) -> Result<(), EventPublisherError> {
        let payload = message
            .encode_to_vec();

        let record = FutureRecord::to(topic)
            .key(key)
            .payload(&payload);

        match self.producer.send(record, self.timeout).await {
            Ok((partition, offset)) => {
                info!(
                    topic = topic,
                    key = key,
                    partition = partition,
                    offset = offset,
                    "Event published successfully"
                );
                Ok(())
            }
            Err((e, _)) => {
                error!(topic = topic, key = key, error = %e, "Failed to publish event");
                Err(EventPublisherError::PublishError(e.to_string()))
            }
        }
    }
}

#[async_trait]
impl EventPublisher for KafkaEventPublisher {
    // ==================== Package Status Events (consolidated topic) ====================

    #[instrument(skip(self, event), fields(package_id = %event.package_id, event_type = "accepted"))]
    async fn publish_package_accepted(
        &self,
        event: PackageAcceptedEvent,
    ) -> Result<(), EventPublisherError> {
        self.publish(topics::PACKAGE_STATUS, &event.package_id, &event)
            .await
    }

    #[instrument(skip(self, event), fields(package_id = %event.package_id, event_type = "in_transit"))]
    async fn publish_package_in_transit(
        &self,
        event: PackageInTransitEvent,
    ) -> Result<(), EventPublisherError> {
        self.publish(topics::PACKAGE_STATUS, &event.package_id, &event)
            .await
    }

    #[instrument(skip(self, event), fields(package_id = %event.package_id, event_type = "delivered"))]
    async fn publish_package_delivered(
        &self,
        event: PackageDeliveredEvent,
    ) -> Result<(), EventPublisherError> {
        self.publish(topics::PACKAGE_STATUS, &event.package_id, &event)
            .await
    }

    #[instrument(skip(self, event), fields(package_id = %event.package_id, event_type = "not_delivered"))]
    async fn publish_package_not_delivered(
        &self,
        event: PackageNotDeliveredEvent,
    ) -> Result<(), EventPublisherError> {
        self.publish(topics::PACKAGE_STATUS, &event.package_id, &event)
            .await
    }

    #[instrument(skip(self, event), fields(package_id = %event.package_id, event_type = "requires_handling"))]
    async fn publish_package_requires_handling(
        &self,
        event: PackageRequiresHandlingEvent,
    ) -> Result<(), EventPublisherError> {
        self.publish(topics::PACKAGE_STATUS, &event.package_id, &event)
            .await
    }

    // ==================== Order Assignment (separate topic) ====================

    #[instrument(skip(self, event), fields(package_id = %event.package_id, courier_id = %event.courier_id))]
    async fn publish_package_assigned(
        &self,
        event: PackageAssignedEvent,
    ) -> Result<(), EventPublisherError> {
        self.publish(topics::ORDER_ASSIGNED, &event.package_id, &event)
            .await
    }

    // ==================== Courier Events (consolidated topic) ====================

    #[instrument(skip(self, event), fields(courier_id = %event.courier_id, event_type = "registered"))]
    async fn publish_courier_registered(
        &self,
        event: CourierRegisteredEvent,
    ) -> Result<(), EventPublisherError> {
        self.publish(topics::COURIER_EVENTS, &event.courier_id, &event)
            .await
    }

    #[instrument(skip(self, event), fields(courier_id = %event.courier_id, event_type = "status_changed"))]
    async fn publish_courier_status_changed(
        &self,
        event: CourierStatusChangedEvent,
    ) -> Result<(), EventPublisherError> {
        self.publish(topics::COURIER_EVENTS, &event.courier_id, &event)
            .await
    }

    // ==================== Courier Location (separate high-frequency topic) ====================

    #[instrument(skip(self, event), fields(courier_id = %event.courier_id))]
    async fn publish_courier_location_updated(
        &self,
        event: CourierLocationUpdatedEvent,
    ) -> Result<(), EventPublisherError> {
        self.publish(topics::COURIER_LOCATION, &event.courier_id, &event)
            .await
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_config_default() {
        let config = KafkaPublisherConfig::default();
        assert_eq!(config.brokers, "localhost:9092");
        assert_eq!(config.client_id, "delivery-service");
    }
}
