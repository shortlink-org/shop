//! Kafka Event Publisher
//!
//! Implementation of EventPublisher using rdkafka.
//! Publishes domain events to Kafka topics.

use async_trait::async_trait;
use chrono::{DateTime, Utc};
use prost::Message;
use rdkafka::message::{Header, OwnedHeaders};
use rdkafka::producer::{FutureProducer, FutureRecord};
use rdkafka::ClientConfig;
use serde::Serialize;
use std::time::Duration;
use tracing::{error, info, instrument};

use crate::domain::model::domain::delivery::common::v1::{
    Address as ProtoAddress, DeliveryPeriod as ProtoDeliveryPeriod,
};
use crate::domain::model::domain::delivery::events::v1::PackageAssignedEvent;
use crate::domain::ports::{DomainEvent, EventPublisher, EventPublisherError};

/// Kafka topic names (consolidated)
/// Format: {domain}.{entity}.{event_group}.v1
///
/// Topics are consolidated by entity lifecycle:
/// - package.status: all package status changes (accepted, assigned, in_transit, delivered, not_delivered, requires_handling)
/// - order.assigned: package assignment to courier for courier-emulation (JSON payload)
/// - courier.events: courier registration and status changes
/// - courier.location: high-frequency location updates (separate for scalability)
pub mod topics {
    /// All package status change events
    /// Events: PackageAccepted, PackageAssigned, PackageInTransit, PackageDelivered, PackageNotDelivered, PackageRequiresHandling
    pub const PACKAGE_STATUS: &str = "delivery.package.status.v1";

    /// Package assigned to courier
    /// Kept separate for courier-emulation compatibility (JSON payload)
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
            brokers: std::env::var("KAFKA_BROKERS")
                .unwrap_or_else(|_| "localhost:9092".to_string()),
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

#[derive(Debug, Clone, PartialEq, Eq)]
pub(crate) enum KafkaPayloadEncoding {
    Proto,
    Json,
}

impl KafkaPayloadEncoding {
    pub(crate) fn as_str(&self) -> &'static str {
        match self {
            Self::Proto => "protobuf",
            Self::Json => "json",
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub(crate) struct KafkaOutboundRecord {
    pub topic: String,
    pub key: String,
    pub payload: Vec<u8>,
    pub headers: Vec<(String, String)>,
    pub payload_encoding: KafkaPayloadEncoding,
    pub event_type: String,
    pub aggregate_id: String,
}

#[derive(Debug, Serialize)]
struct OrderAssignedAddressPayload {
    street: String,
    city: String,
    postal_code: String,
    country: String,
    latitude: f64,
    longitude: f64,
}

#[derive(Debug, Serialize)]
struct OrderAssignedPeriodPayload {
    start_time: DateTime<Utc>,
    end_time: DateTime<Utc>,
}

#[derive(Debug, Serialize)]
struct OrderAssignedPayload {
    package_id: String,
    courier_id: String,
    status: i32,
    assigned_at: DateTime<Utc>,
    pickup_address: OrderAssignedAddressPayload,
    delivery_address: OrderAssignedAddressPayload,
    delivery_period: OrderAssignedPeriodPayload,
    occurred_at: DateTime<Utc>,
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

    pub(crate) async fn publish_record(
        &self,
        record: &KafkaOutboundRecord,
    ) -> Result<(), EventPublisherError> {
        let mut kafka_record = FutureRecord::to(&record.topic)
            .key(record.key.as_str())
            .payload(&record.payload);

        if !record.headers.is_empty() {
            let headers =
                record
                    .headers
                    .iter()
                    .fold(OwnedHeaders::new(), |headers, (key, value)| {
                        headers.insert(Header {
                            key: key.as_str(),
                            value: Some(value.as_str()),
                        })
                    });
            kafka_record = kafka_record.headers(headers);
        }

        match self.producer.send(kafka_record, self.timeout).await {
            Ok(delivery) => {
                info!(
                    topic = record.topic,
                    key = record.key,
                    partition = delivery.partition,
                    offset = delivery.offset,
                    "Event published successfully"
                );
                Ok(())
            }
            Err((e, _)) => {
                error!(topic = record.topic, key = record.key, error = %e, "Failed to publish event");
                Err(EventPublisherError::PublishFailed(e.to_string()))
            }
        }
    }

    fn build_proto_record<M: Message>(
        topic: &str,
        key: &str,
        event_type: &str,
        message: &M,
    ) -> KafkaOutboundRecord {
        KafkaOutboundRecord {
            topic: topic.to_string(),
            key: key.to_string(),
            payload: message.encode_to_vec(),
            headers: vec![("event_type".to_string(), event_type.to_string())],
            payload_encoding: KafkaPayloadEncoding::Proto,
            event_type: event_type.to_string(),
            aggregate_id: key.to_string(),
        }
    }

    fn build_json_record<P: Serialize>(
        topic: &str,
        key: &str,
        event_type: &str,
        payload: &P,
    ) -> Result<KafkaOutboundRecord, EventPublisherError> {
        Ok(KafkaOutboundRecord {
            topic: topic.to_string(),
            key: key.to_string(),
            payload: serde_json::to_vec(payload)
                .map_err(|e| EventPublisherError::SerializationFailed(e.to_string()))?,
            headers: Vec::new(),
            payload_encoding: KafkaPayloadEncoding::Json,
            event_type: event_type.to_string(),
            aggregate_id: key.to_string(),
        })
    }

    fn timestamp_to_datetime(
        timestamp: &Option<pbjson_types::Timestamp>,
        field: &str,
    ) -> Result<DateTime<Utc>, EventPublisherError> {
        let timestamp = timestamp.as_ref().ok_or_else(|| {
            EventPublisherError::SerializationFailed(format!("{field} is required"))
        })?;

        DateTime::from_timestamp(timestamp.seconds, timestamp.nanos as u32)
            .ok_or_else(|| EventPublisherError::SerializationFailed(format!("{field} is invalid")))
    }

    fn map_address(
        address: &Option<ProtoAddress>,
        field: &str,
    ) -> Result<OrderAssignedAddressPayload, EventPublisherError> {
        let address = address.as_ref().ok_or_else(|| {
            EventPublisherError::SerializationFailed(format!("{field} is required"))
        })?;

        Ok(OrderAssignedAddressPayload {
            street: address.street.clone(),
            city: address.city.clone(),
            postal_code: address.postal_code.clone(),
            country: address.country.clone(),
            latitude: address.latitude,
            longitude: address.longitude,
        })
    }

    fn map_delivery_period(
        period: &Option<ProtoDeliveryPeriod>,
    ) -> Result<OrderAssignedPeriodPayload, EventPublisherError> {
        let period = period.as_ref().ok_or_else(|| {
            EventPublisherError::SerializationFailed("delivery_period is required".to_string())
        })?;

        Ok(OrderAssignedPeriodPayload {
            start_time: Self::timestamp_to_datetime(
                &period.start_time,
                "delivery_period.start_time",
            )?,
            end_time: Self::timestamp_to_datetime(&period.end_time, "delivery_period.end_time")?,
        })
    }

    fn map_assigned_payload(
        event: &PackageAssignedEvent,
    ) -> Result<OrderAssignedPayload, EventPublisherError> {
        Ok(OrderAssignedPayload {
            package_id: event.package_id.clone(),
            courier_id: event.courier_id.clone(),
            status: event.status,
            assigned_at: Self::timestamp_to_datetime(&event.assigned_at, "assigned_at")?,
            pickup_address: Self::map_address(&event.pickup_address, "pickup_address")?,
            delivery_address: Self::map_address(&event.delivery_address, "delivery_address")?,
            delivery_period: Self::map_delivery_period(&event.delivery_period)?,
            occurred_at: Self::timestamp_to_datetime(&event.occurred_at, "occurred_at")?,
        })
    }

    pub(crate) fn encode_event(
        event: &DomainEvent,
    ) -> Result<Vec<KafkaOutboundRecord>, EventPublisherError> {
        match event {
            DomainEvent::PackageAccepted(e) => Ok(vec![Self::build_proto_record(
                topics::PACKAGE_STATUS,
                &e.package_id,
                "PackageAcceptedEvent",
                e,
            )]),
            DomainEvent::PackageInTransit(e) => Ok(vec![Self::build_proto_record(
                topics::PACKAGE_STATUS,
                &e.package_id,
                "PackageInTransitEvent",
                e,
            )]),
            DomainEvent::PackageDelivered(e) => Ok(vec![Self::build_proto_record(
                topics::PACKAGE_STATUS,
                &e.package_id,
                "PackageDeliveredEvent",
                e,
            )]),
            DomainEvent::PackageNotDelivered(e) => Ok(vec![Self::build_proto_record(
                topics::PACKAGE_STATUS,
                &e.package_id,
                "PackageNotDeliveredEvent",
                e,
            )]),
            DomainEvent::PackageRequiresHandling(e) => Ok(vec![Self::build_proto_record(
                topics::PACKAGE_STATUS,
                &e.package_id,
                "PackageRequiresHandlingEvent",
                e,
            )]),
            DomainEvent::PackageAssigned(e) => {
                let payload = Self::map_assigned_payload(e)?;
                Ok(vec![
                    Self::build_proto_record(
                        topics::PACKAGE_STATUS,
                        &e.package_id,
                        "PackageAssignedEvent",
                        e,
                    ),
                    Self::build_json_record(
                        topics::ORDER_ASSIGNED,
                        &e.package_id,
                        "PackageAssignedEvent",
                        &payload,
                    )?,
                ])
            }
            DomainEvent::CourierRegistered(e) => Ok(vec![Self::build_proto_record(
                topics::COURIER_EVENTS,
                &e.courier_id,
                "CourierRegisteredEvent",
                e,
            )]),
            DomainEvent::CourierStatusChanged(e) => Ok(vec![Self::build_proto_record(
                topics::COURIER_EVENTS,
                &e.courier_id,
                "CourierStatusChangedEvent",
                e,
            )]),
            DomainEvent::CourierLocationUpdated(e) => Ok(vec![Self::build_proto_record(
                topics::COURIER_LOCATION,
                &e.courier_id,
                "CourierLocationUpdatedEvent",
                e,
            )]),
        }
    }
}

#[async_trait]
impl EventPublisher for KafkaEventPublisher {
    #[instrument(skip(self, event))]
    async fn publish(&self, event: DomainEvent) -> Result<(), EventPublisherError> {
        let records = Self::encode_event(&event)?;
        for record in &records {
            self.publish_record(record).await?;
        }
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::domain::delivery::common::v1::{Address, DeliveryPeriod};
    use crate::domain::model::domain::delivery::events::v1::PackageAssignedEvent;

    #[test]
    fn test_config_default() {
        let config = KafkaPublisherConfig::default();
        assert_eq!(config.brokers, "localhost:9092");
        assert_eq!(config.client_id, "delivery-service");
    }

    #[test]
    fn test_map_assigned_payload() {
        let event = PackageAssignedEvent {
            package_id: "pkg-1".to_string(),
            courier_id: "courier-1".to_string(),
            status: 3,
            assigned_at: Some(pbjson_types::Timestamp {
                seconds: 1_762_858_800,
                nanos: 0,
            }),
            pickup_address: Some(Address {
                street: "Pickup".to_string(),
                city: "Berlin".to_string(),
                postal_code: "10115".to_string(),
                country: String::new(),
                latitude: 52.52,
                longitude: 13.40,
            }),
            delivery_address: Some(Address {
                street: "Dropoff".to_string(),
                city: "Berlin".to_string(),
                postal_code: "10117".to_string(),
                country: String::new(),
                latitude: 52.53,
                longitude: 13.41,
            }),
            delivery_period: Some(DeliveryPeriod {
                start_time: Some(pbjson_types::Timestamp {
                    seconds: 1_762_859_160,
                    nanos: 0,
                }),
                end_time: Some(pbjson_types::Timestamp {
                    seconds: 1_762_862_760,
                    nanos: 0,
                }),
            }),
            customer_phone: String::new(),
            occurred_at: Some(pbjson_types::Timestamp {
                seconds: 1_762_858_810,
                nanos: 0,
            }),
        };

        let payload = KafkaEventPublisher::map_assigned_payload(&event).expect("payload");

        assert_eq!(payload.package_id, "pkg-1");
        assert_eq!(payload.courier_id, "courier-1");
        assert_eq!(payload.pickup_address.city, "Berlin");
        assert_eq!(payload.delivery_address.postal_code, "10117");
    }

    #[test]
    fn test_map_assigned_payload_requires_timestamps() {
        let event = PackageAssignedEvent {
            package_id: "pkg-1".to_string(),
            courier_id: "courier-1".to_string(),
            status: 3,
            assigned_at: None,
            pickup_address: None,
            delivery_address: None,
            delivery_period: None,
            customer_phone: String::new(),
            occurred_at: None,
        };

        let err = KafkaEventPublisher::map_assigned_payload(&event).expect_err("expected error");

        assert!(matches!(err, EventPublisherError::SerializationFailed(_)));
    }

    #[test]
    fn test_encode_event_for_assigned_creates_two_records() {
        let event = PackageAssignedEvent {
            package_id: "pkg-1".to_string(),
            courier_id: "courier-1".to_string(),
            status: 3,
            assigned_at: Some(pbjson_types::Timestamp {
                seconds: 1_762_858_800,
                nanos: 0,
            }),
            pickup_address: Some(Address {
                street: "Pickup".to_string(),
                city: "Berlin".to_string(),
                postal_code: "10115".to_string(),
                country: String::new(),
                latitude: 52.52,
                longitude: 13.40,
            }),
            delivery_address: Some(Address {
                street: "Dropoff".to_string(),
                city: "Berlin".to_string(),
                postal_code: "10117".to_string(),
                country: String::new(),
                latitude: 52.53,
                longitude: 13.41,
            }),
            delivery_period: Some(DeliveryPeriod {
                start_time: Some(pbjson_types::Timestamp {
                    seconds: 1_762_859_160,
                    nanos: 0,
                }),
                end_time: Some(pbjson_types::Timestamp {
                    seconds: 1_762_862_760,
                    nanos: 0,
                }),
            }),
            customer_phone: String::new(),
            occurred_at: Some(pbjson_types::Timestamp {
                seconds: 1_762_858_810,
                nanos: 0,
            }),
        };

        let records =
            KafkaEventPublisher::encode_event(&DomainEvent::PackageAssigned(event)).unwrap();

        assert_eq!(records.len(), 2);
        assert_eq!(records[0].topic, topics::PACKAGE_STATUS);
        assert_eq!(records[0].payload_encoding, KafkaPayloadEncoding::Proto);
        assert_eq!(records[1].topic, topics::ORDER_ASSIGNED);
        assert_eq!(records[1].payload_encoding, KafkaPayloadEncoding::Json);
    }
}
