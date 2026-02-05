//! Kafka Event Publisher
//!
//! Implementation of EventPublisher using rdkafka.
//! Publishes domain events to Kafka topics.

use async_trait::async_trait;
use prost::Message;
use rdkafka::message::{Header, OwnedHeaders};
use rdkafka::producer::{FutureProducer, FutureRecord};
use rdkafka::ClientConfig;
use std::time::Duration;
use tracing::{error, info, instrument};

use crate::domain::ports::{DomainEvent, EventPublisher, EventPublisherError};

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
            .map_err(|e| EventPublisherError::PublishFailed(e.to_string()))?;

        info!("Kafka publisher connected to {}", config.brokers);

        Ok(Self {
            producer,
            timeout: Duration::from_millis(config.message_timeout_ms),
        })
    }

    /// Publish a protobuf message to a topic with an event_type header for consumer dispatch.
    async fn publish_to_topic<M: Message>(
        &self,
        topic: &str,
        key: &str,
        event_type: &str,
        message: &M,
    ) -> Result<(), EventPublisherError> {
        let payload = message.encode_to_vec();
        let headers = OwnedHeaders::new().insert(Header {
            key: "event_type",
            value: Some(event_type),
        });

        let record = FutureRecord::to(topic)
            .key(key)
            .payload(&payload)
            .headers(headers);

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
                Err(EventPublisherError::PublishFailed(e.to_string()))
            }
        }
    }
}

#[async_trait]
impl EventPublisher for KafkaEventPublisher {
    #[instrument(skip(self, event))]
    async fn publish(&self, event: DomainEvent) -> Result<(), EventPublisherError> {
        match event {
            DomainEvent::PackageAccepted(e) => {
                self.publish_to_topic(
                    topics::PACKAGE_STATUS,
                    &e.package_id,
                    "PackageAcceptedEvent",
                    &e,
                )
                .await
            }
            DomainEvent::PackageInTransit(e) => {
                self.publish_to_topic(
                    topics::PACKAGE_STATUS,
                    &e.package_id,
                    "PackageInTransitEvent",
                    &e,
                )
                .await
            }
            DomainEvent::PackageDelivered(e) => {
                self.publish_to_topic(
                    topics::PACKAGE_STATUS,
                    &e.package_id,
                    "PackageDeliveredEvent",
                    &e,
                )
                .await
            }
            DomainEvent::PackageNotDelivered(e) => {
                self.publish_to_topic(
                    topics::PACKAGE_STATUS,
                    &e.package_id,
                    "PackageNotDeliveredEvent",
                    &e,
                )
                .await
            }
            DomainEvent::PackageRequiresHandling(e) => {
                self.publish_to_topic(
                    topics::PACKAGE_STATUS,
                    &e.package_id,
                    "PackageRequiresHandlingEvent",
                    &e,
                )
                .await
            }
            DomainEvent::PackageAssigned(e) => {
                self.publish_to_topic(
                    topics::ORDER_ASSIGNED,
                    &e.package_id,
                    "PackageAssignedEvent",
                    &e,
                )
                .await
            }
            DomainEvent::CourierRegistered(e) => {
                self.publish_to_topic(
                    topics::COURIER_EVENTS,
                    &e.courier_id,
                    "CourierRegisteredEvent",
                    &e,
                )
                .await
            }
            DomainEvent::CourierStatusChanged(e) => {
                self.publish_to_topic(
                    topics::COURIER_EVENTS,
                    &e.courier_id,
                    "CourierStatusChangedEvent",
                    &e,
                )
                .await
            }
            DomainEvent::CourierLocationUpdated(e) => {
                self.publish_to_topic(
                    topics::COURIER_LOCATION,
                    &e.courier_id,
                    "CourierLocationUpdatedEvent",
                    &e,
                )
                .await
            }
        }
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
