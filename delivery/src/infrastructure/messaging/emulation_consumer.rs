//! Kafka consumer for courier-emulation delivery lifecycle events.
//!
//! Consumes pickup and delivery confirmation events emitted by courier-emulation
//! and applies the corresponding package commands in the Delivery service.

use std::sync::Arc;

use rdkafka::consumer::{Consumer, StreamConsumer};
use rdkafka::message::Message;
use rdkafka::ClientConfig;
use serde::Deserialize;
use tokio::sync::broadcast;
use tracing::{error, info};
use uuid::Uuid;

use crate::domain::model::package::PackageId;
use crate::domain::model::package::PackageStatus;
use crate::domain::model::vo::location::Location;
use crate::domain::ports::{
    CommandHandlerWithResult, CourierCache, CourierRepository, GeolocationService,
    PackageRepository,
};
use crate::infrastructure::messaging::RedisTrackingPubSub;
use crate::usecases::package::command::{
    deliver_order, pick_up_order, DeliverOrderError, DeliverOrderHandler, PickUpOrderError,
    PickUpOrderHandler,
};

/// Kafka topic for pickup confirmations from courier-emulation.
pub const TOPIC_PACKAGE_PICKED_UP: &str = "delivery.order.order_picked_up.v1";

/// Kafka topic for delivery confirmations from courier-emulation.
pub const TOPIC_PACKAGE_DELIVERED: &str = "delivery.order.order_delivered.v1";

/// Consumer group ID for emulation events.
pub const CONSUMER_GROUP: &str = "delivery-service";

const DEFAULT_ACCURACY_METERS: f64 = 10.0;

#[derive(Debug, Clone, Deserialize)]
struct EmulationLocation {
    latitude: f64,
    longitude: f64,
    #[serde(default = "default_accuracy")]
    accuracy: f64,
}

#[derive(Debug, Clone, Deserialize)]
struct PickUpPackageEvent {
    #[serde(alias = "order_id")]
    package_id: String,
    courier_id: String,
    pickup_location: EmulationLocation,
}

#[derive(Debug, Clone, Deserialize)]
struct DeliverPackageEvent {
    #[serde(alias = "order_id")]
    package_id: String,
    courier_id: String,
    status: String,
    #[serde(default)]
    reason: Option<String>,
    current_location: EmulationLocation,
}

fn default_accuracy() -> f64 {
    DEFAULT_ACCURACY_METERS
}

/// Configuration for emulation events consumer.
#[derive(Debug, Clone)]
pub struct EmulationConsumerConfig {
    /// Kafka bootstrap servers.
    pub brokers: String,
    /// Consumer group ID.
    pub group_id: String,
    /// Topic for pickup confirmations.
    pub pickup_topic: String,
    /// Topic for delivery confirmations.
    pub delivery_topic: String,
}

impl Default for EmulationConsumerConfig {
    fn default() -> Self {
        Self {
            brokers: "localhost:9092".to_string(),
            group_id: CONSUMER_GROUP.to_string(),
            pickup_topic: TOPIC_PACKAGE_PICKED_UP.to_string(),
            delivery_topic: TOPIC_PACKAGE_DELIVERED.to_string(),
        }
    }
}

impl EmulationConsumerConfig {
    /// Create config from environment variables.
    pub fn from_env() -> Self {
        Self {
            brokers: std::env::var("KAFKA_BROKERS")
                .unwrap_or_else(|_| "localhost:9092".to_string()),
            group_id: std::env::var("KAFKA_CONSUMER_GROUP")
                .unwrap_or_else(|_| CONSUMER_GROUP.to_string()),
            pickup_topic: std::env::var("KAFKA_PICKUP_TOPIC")
                .unwrap_or_else(|_| TOPIC_PACKAGE_PICKED_UP.to_string()),
            delivery_topic: std::env::var("KAFKA_DELIVERY_RESULT_TOPIC")
                .unwrap_or_else(|_| TOPIC_PACKAGE_DELIVERED.to_string()),
        }
    }
}

/// Consumes courier-emulation pickup and delivery confirmations.
pub struct EmulationConsumer<R, C, P, G>
where
    R: CourierRepository,
    C: CourierCache,
    P: PackageRepository,
    G: GeolocationService,
{
    consumer: StreamConsumer,
    courier_repo: Arc<R>,
    courier_cache: Arc<C>,
    package_repo: Arc<P>,
    geolocation_service: Arc<G>,
    tracking_pubsub: Arc<RedisTrackingPubSub>,
    config: EmulationConsumerConfig,
    shutdown_rx: broadcast::Receiver<()>,
}

impl<R, C, P, G> EmulationConsumer<R, C, P, G>
where
    R: CourierRepository + Send + Sync + 'static,
    C: CourierCache + Send + Sync + 'static,
    P: PackageRepository + Send + Sync + 'static,
    G: GeolocationService + Send + Sync + 'static,
{
    /// Create a new emulation event consumer.
    pub fn new(
        config: EmulationConsumerConfig,
        courier_repo: Arc<R>,
        courier_cache: Arc<C>,
        package_repo: Arc<P>,
        geolocation_service: Arc<G>,
        tracking_pubsub: Arc<RedisTrackingPubSub>,
        shutdown_rx: broadcast::Receiver<()>,
    ) -> Result<Self, String> {
        let consumer: StreamConsumer = ClientConfig::new()
            .set("bootstrap.servers", &config.brokers)
            .set("group.id", &config.group_id)
            .set("enable.auto.commit", "true")
            .set("auto.commit.interval.ms", "5000")
            .set("auto.offset.reset", "latest")
            .set("session.timeout.ms", "30000")
            .create()
            .map_err(|e| format!("Failed to create Kafka consumer: {}", e))?;

        consumer
            .subscribe(&[&config.pickup_topic, &config.delivery_topic])
            .map_err(|e| format!("Failed to subscribe to emulation topics: {}", e))?;

        info!(
            pickup_topic = %config.pickup_topic,
            delivery_topic = %config.delivery_topic,
            "Emulation consumer subscribed to topics"
        );

        Ok(Self {
            consumer,
            courier_repo,
            courier_cache,
            package_repo,
            geolocation_service,
            tracking_pubsub,
            config,
            shutdown_rx,
        })
    }

    /// Run the consumer loop.
    pub async fn run(mut self) {
        info!(
            pickup_topic = %self.config.pickup_topic,
            delivery_topic = %self.config.delivery_topic,
            "Starting emulation consumer"
        );

        loop {
            tokio::select! {
                _ = self.shutdown_rx.recv() => {
                    info!("Emulation consumer received shutdown signal");
                    break;
                }
                message = self.consumer.recv() => {
                    match message {
                        Ok(msg) => {
                            if let Some(payload) = msg.payload() {
                                if let Err(err) = self.process_message(msg.topic(), payload).await {
                                    error!(topic = msg.topic(), error = %err, "Failed to process emulation event");
                                }
                            }
                        }
                        Err(err) => {
                            error!(error = %err, "Error receiving emulation event from Kafka");
                        }
                    }
                }
            }
        }

        info!("Emulation consumer stopped");
    }

    async fn process_message(&self, topic: &str, payload: &[u8]) -> Result<(), String> {
        if topic == self.config.pickup_topic {
            let event: PickUpPackageEvent = serde_json::from_slice(payload)
                .map_err(|e| format!("Failed to deserialize pickup event: {}", e))?;
            return self.handle_pickup(event).await;
        }

        if topic == self.config.delivery_topic {
            let event: DeliverPackageEvent = serde_json::from_slice(payload)
                .map_err(|e| format!("Failed to deserialize delivery event: {}", e))?;
            return self.handle_delivery(event).await;
        }

        Err(format!("Unsupported emulation topic: {topic}"))
    }

    async fn handle_pickup(&self, event: PickUpPackageEvent) -> Result<(), String> {
        let package_id = parse_uuid("package_id", &event.package_id)?;
        let courier_id = parse_uuid("courier_id", &event.courier_id)?;
        let pickup_location = to_domain_location(&event.pickup_location)?;

        let handler =
            PickUpOrderHandler::new(self.package_repo.clone(), self.geolocation_service.clone());

        match handler
            .handle(pick_up_order::Command::new(
                package_id,
                courier_id,
                pickup_location,
            ))
            .await
        {
            Ok(_) => {
                self.notify_tracking_for_package(package_id).await;
                info!(package_id = %package_id, courier_id = %courier_id, "Processed pickup event from courier-emulation");
                Ok(())
            }
            Err(PickUpOrderError::AlreadyPickedUp(_)) => {
                info!(package_id = %package_id, courier_id = %courier_id, "Ignoring duplicate pickup event from courier-emulation");
                Ok(())
            }
            Err(PickUpOrderError::InvalidPackageStatus(
                PackageStatus::Delivered | PackageStatus::NotDelivered | PackageStatus::InTransit,
            )) => {
                info!(package_id = %package_id, courier_id = %courier_id, "Ignoring stale pickup event from courier-emulation");
                Ok(())
            }
            Err(err) => Err(format!("pickup command failed: {err}")),
        }
    }

    async fn handle_delivery(&self, event: DeliverPackageEvent) -> Result<(), String> {
        let package_id = parse_uuid("package_id", &event.package_id)?;
        let courier_id = parse_uuid("courier_id", &event.courier_id)?;
        let confirmation_location = to_domain_location(&event.current_location)?;

        let cmd = match event.status.as_str() {
            "DELIVERED" => deliver_order::Command::delivered(
                package_id,
                courier_id,
                confirmation_location,
                None,
                None,
            ),
            "NOT_DELIVERED" => deliver_order::Command::not_delivered(
                package_id,
                courier_id,
                confirmation_location,
                map_not_delivered_reason(event.reason.as_deref())?,
                None,
            ),
            status => return Err(format!("Unsupported delivery status: {status}")),
        };

        let handler = DeliverOrderHandler::new(
            self.courier_repo.clone(),
            self.courier_cache.clone(),
            self.package_repo.clone(),
            self.geolocation_service.clone(),
        );

        match handler.handle(cmd).await {
            Ok(_) => {
                self.notify_tracking_for_package(package_id).await;
                info!(package_id = %package_id, courier_id = %courier_id, "Processed delivery event from courier-emulation");
                Ok(())
            }
            Err(DeliverOrderError::AlreadyDelivered(_)) => {
                info!(package_id = %package_id, courier_id = %courier_id, "Ignoring duplicate delivered event from courier-emulation");
                Ok(())
            }
            Err(DeliverOrderError::InvalidPackageStatus(
                PackageStatus::Delivered | PackageStatus::NotDelivered,
            )) => {
                info!(package_id = %package_id, courier_id = %courier_id, "Ignoring stale delivery event from courier-emulation");
                Ok(())
            }
            Err(err) => Err(format!("delivery command failed: {err}")),
        }
    }

    async fn notify_tracking_for_package(&self, package_id: Uuid) {
        if let Ok(Some(package)) = self
            .package_repo
            .find_by_id(PackageId::from_uuid(package_id))
            .await
        {
            if let Err(err) = self
                .tracking_pubsub
                .publish_order_update(package.order_id())
                .await
            {
                error!(order_id = %package.order_id(), error = %err, "Failed to publish tracking update");
            }
        }
    }
}

fn parse_uuid(field: &str, value: &str) -> Result<Uuid, String> {
    Uuid::parse_str(value).map_err(|e| format!("Invalid {field} '{value}': {e}"))
}

fn to_domain_location(location: &EmulationLocation) -> Result<Location, String> {
    Location::new(location.latitude, location.longitude, location.accuracy)
        .map_err(|e| format!("Invalid location: {}", e))
}

fn map_not_delivered_reason(
    reason: Option<&str>,
) -> Result<deliver_order::NotDeliveredReason, String> {
    match reason {
        Some("CUSTOMER_NOT_AVAILABLE") => {
            Ok(deliver_order::NotDeliveredReason::CustomerUnavailable)
        }
        Some("WRONG_ADDRESS") => Ok(deliver_order::NotDeliveredReason::WrongAddress),
        Some("CUSTOMER_REFUSED") => Ok(deliver_order::NotDeliveredReason::Refused),
        Some("ACCESS_DENIED") => Ok(deliver_order::NotDeliveredReason::AccessDenied),
        Some("OTHER") => Ok(deliver_order::NotDeliveredReason::Other(
            "courier-emulation".to_string(),
        )),
        Some(value) => Err(format!("Unsupported not delivered reason: {value}")),
        None => Err("Missing not delivered reason".to_string()),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_deserialize_pickup_event_supports_package_id() {
        let payload = r#"{
            "package_id": "pkg-1",
            "courier_id": "courier-1",
            "pickup_location": {
                "latitude": 52.52,
                "longitude": 13.405,
                "accuracy": 10.0
            }
        }"#;

        let event: PickUpPackageEvent = serde_json::from_str(payload).expect("pickup event");

        assert_eq!(event.package_id, "pkg-1");
        assert_eq!(event.courier_id, "courier-1");
        assert_eq!(event.pickup_location.accuracy, 10.0);
    }

    #[test]
    fn test_deserialize_pickup_event_supports_legacy_order_id_alias() {
        let payload = r#"{
            "order_id": "pkg-legacy",
            "courier_id": "courier-1",
            "pickup_location": {
                "latitude": 52.52,
                "longitude": 13.405
            }
        }"#;

        let event: PickUpPackageEvent = serde_json::from_str(payload).expect("pickup event");

        assert_eq!(event.package_id, "pkg-legacy");
        assert_eq!(event.pickup_location.accuracy, DEFAULT_ACCURACY_METERS);
    }

    #[test]
    fn test_map_not_delivered_reason() {
        assert!(matches!(
            map_not_delivered_reason(Some("CUSTOMER_REFUSED")),
            Ok(deliver_order::NotDeliveredReason::Refused)
        ));
        assert!(map_not_delivered_reason(Some("INVALID")).is_err());
        assert!(map_not_delivered_reason(None).is_err());
    }
}
