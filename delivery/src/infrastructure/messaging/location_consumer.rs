//! Kafka Location Consumer
//!
//! Consumes courier location updates from Kafka and stores them in cache and database.

use rdkafka::consumer::{Consumer, StreamConsumer};
use rdkafka::message::Message;
use rdkafka::{ClientConfig, TopicPartitionList};
use serde::Deserialize;
use std::sync::Arc;
use tokio::sync::broadcast;
use tracing::{error, info};
use uuid::Uuid;

use crate::domain::model::courier_location::{LocationHistoryEntry};
use crate::domain::model::vo::location::Location;
use crate::domain::model::CourierLocation;
use crate::domain::ports::{LocationCache, LocationRepository};

/// Kafka topic for courier location updates
pub const TOPIC_COURIER_LOCATION_UPDATES: &str = "courier.location.updates";

/// Consumer group ID
pub const CONSUMER_GROUP: &str = "delivery-service";

/// Courier location event from courier-emulation service
#[derive(Debug, Clone, Deserialize)]
pub struct CourierLocationEvent {
    /// Courier ID
    pub courier_id: String,
    /// Latitude
    pub latitude: f64,
    /// Longitude
    pub longitude: f64,
    /// Accuracy in meters (optional)
    #[serde(default = "default_accuracy")]
    pub accuracy: f64,
    /// Timestamp (ISO 8601 string or unix timestamp)
    pub timestamp: String,
    /// Speed in km/h
    #[serde(default)]
    pub speed: Option<f64>,
    /// Heading in degrees (0-360)
    #[serde(default)]
    pub heading: Option<f64>,
    /// Route ID (optional)
    #[serde(default)]
    pub route_id: Option<String>,
    /// Status (moving, idle, delivering)
    #[serde(default)]
    pub status: Option<String>,
}

fn default_accuracy() -> f64 {
    10.0 // Default accuracy of 10 meters
}

/// Configuration for location consumer
#[derive(Debug, Clone)]
pub struct LocationConsumerConfig {
    /// Kafka bootstrap servers
    pub brokers: String,
    /// Consumer group ID
    pub group_id: String,
    /// Topic to consume from
    pub topic: String,
}

impl Default for LocationConsumerConfig {
    fn default() -> Self {
        Self {
            brokers: "localhost:9092".to_string(),
            group_id: CONSUMER_GROUP.to_string(),
            topic: TOPIC_COURIER_LOCATION_UPDATES.to_string(),
        }
    }
}

impl LocationConsumerConfig {
    /// Create config from environment variables
    pub fn from_env() -> Self {
        Self {
            brokers: std::env::var("KAFKA_BROKERS").unwrap_or_else(|_| "localhost:9092".to_string()),
            group_id: std::env::var("KAFKA_CONSUMER_GROUP")
                .unwrap_or_else(|_| CONSUMER_GROUP.to_string()),
            topic: std::env::var("KAFKA_LOCATION_TOPIC")
                .unwrap_or_else(|_| TOPIC_COURIER_LOCATION_UPDATES.to_string()),
        }
    }
}

/// Location consumer for processing courier location updates
pub struct LocationConsumer<C: LocationCache, R: LocationRepository> {
    consumer: StreamConsumer,
    location_cache: Arc<C>,
    location_repository: Arc<R>,
    config: LocationConsumerConfig,
    shutdown_rx: broadcast::Receiver<()>,
}

impl<C: LocationCache + 'static, R: LocationRepository + 'static> LocationConsumer<C, R> {
    /// Create a new location consumer
    pub fn new(
        config: LocationConsumerConfig,
        location_cache: Arc<C>,
        location_repository: Arc<R>,
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

        // Subscribe to topic
        let mut topics = TopicPartitionList::new();
        topics.add_partition(&config.topic, 0);
        
        consumer
            .subscribe(&[&config.topic])
            .map_err(|e| format!("Failed to subscribe to topic {}: {}", config.topic, e))?;

        info!("Location consumer subscribed to topic: {}", config.topic);

        Ok(Self {
            consumer,
            location_cache,
            location_repository,
            config,
            shutdown_rx,
        })
    }

    /// Run the consumer loop
    pub async fn run(mut self) {
        info!("Starting location consumer for topic: {}", self.config.topic);

        loop {
            tokio::select! {
                _ = self.shutdown_rx.recv() => {
                    info!("Location consumer received shutdown signal");
                    break;
                }
                message = self.consumer.recv() => {
                    match message {
                        Ok(msg) => {
                            if let Some(payload) = msg.payload() {
                                if let Err(e) = self.process_message(payload).await {
                                    error!("Failed to process location message: {}", e);
                                }
                            }
                        }
                        Err(e) => {
                            error!("Error receiving message from Kafka: {}", e);
                        }
                    }
                }
            }
        }

        info!("Location consumer stopped");
    }

    /// Process a single location message
    async fn process_message(&self, payload: &[u8]) -> Result<(), String> {
        // Deserialize the event
        let event: CourierLocationEvent = serde_json::from_slice(payload)
            .map_err(|e| format!("Failed to deserialize location event: {}", e))?;

        // Parse courier ID
        let courier_id = Uuid::parse_str(&event.courier_id)
            .map_err(|e| format!("Invalid courier ID '{}': {}", event.courier_id, e))?;

        // Parse timestamp
        let timestamp = chrono::DateTime::parse_from_rfc3339(&event.timestamp)
            .map(|dt| dt.with_timezone(&chrono::Utc))
            .or_else(|_| {
                // Try parsing as unix timestamp
                event.timestamp.parse::<i64>()
                    .map(|ts| chrono::DateTime::from_timestamp(ts, 0).unwrap_or_else(chrono::Utc::now))
            })
            .unwrap_or_else(|_| chrono::Utc::now());

        // Create Location value object
        let location = Location::new(event.latitude, event.longitude, event.accuracy)
            .map_err(|e| format!("Invalid location: {}", e))?;

        // Create CourierLocation entity (skip timestamp validation for incoming events)
        let courier_location = CourierLocation::from_stored(
            courier_id,
            location.clone(),
            timestamp,
            event.speed,
            event.heading,
        )
        .map_err(|e| format!("Invalid courier location: {}", e))?;

        // Store in Redis cache (hot data)
        self.location_cache
            .set_location(&courier_location)
            .await
            .map_err(|e| format!("Failed to cache location: {}", e))?;

        // Store in PostgreSQL (history)
        let history_entry = LocationHistoryEntry::new(
            courier_id,
            location,
            timestamp,
            event.speed,
            event.heading,
        );

        self.location_repository
            .save(&history_entry)
            .await
            .map_err(|e| format!("Failed to save location history: {}", e))?;

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_deserialize_location_event() {
        let json = r#"{
            "courier_id": "550e8400-e29b-41d4-a716-446655440000",
            "latitude": 52.52,
            "longitude": 13.405,
            "accuracy": 10.0,
            "timestamp": "2024-01-15T10:30:00Z",
            "speed": 35.5,
            "heading": 180.0,
            "route_id": "route-123",
            "status": "moving"
        }"#;

        let event: CourierLocationEvent = serde_json::from_str(json).unwrap();
        assert_eq!(event.courier_id, "550e8400-e29b-41d4-a716-446655440000");
        assert_eq!(event.latitude, 52.52);
        assert_eq!(event.longitude, 13.405);
        assert_eq!(event.speed, Some(35.5));
        assert_eq!(event.heading, Some(180.0));
    }

    #[test]
    fn test_deserialize_minimal_location_event() {
        let json = r#"{
            "courier_id": "550e8400-e29b-41d4-a716-446655440000",
            "latitude": 52.52,
            "longitude": 13.405,
            "timestamp": "2024-01-15T10:30:00Z"
        }"#;

        let event: CourierLocationEvent = serde_json::from_str(json).unwrap();
        assert_eq!(event.accuracy, 10.0); // default
        assert_eq!(event.speed, None);
        assert_eq!(event.heading, None);
    }
}
