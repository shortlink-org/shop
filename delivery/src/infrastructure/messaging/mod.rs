//! Messaging Infrastructure
//!
//! Contains implementations for message broker integration.
//! - Kafka event publisher for domain events
//! - Kafka consumer for location updates

pub mod kafka_publisher;
pub mod location_consumer;

pub use kafka_publisher::KafkaEventPublisher;
pub use location_consumer::LocationConsumer;
