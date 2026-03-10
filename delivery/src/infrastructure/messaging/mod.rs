//! Messaging Infrastructure
//!
//! Contains implementations for message broker integration.
//! - Kafka event publisher for domain events
//! - Kafka consumer for location updates
//! - Kafka consumer for courier-emulation pickup/delivery confirmations

pub mod emulation_consumer;
pub mod kafka_publisher;
pub mod location_consumer;
pub mod outbox_forwarder;

pub use emulation_consumer::EmulationConsumer;
pub use kafka_publisher::KafkaEventPublisher;
pub use location_consumer::LocationConsumer;
pub use outbox_forwarder::OutboxForwarder;
