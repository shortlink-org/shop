//! Configuration Module
//!
//! Loads configuration from environment variables.

use std::env;

use thiserror::Error;

/// Configuration errors
#[derive(Debug, Error)]
pub enum ConfigError {
    #[error("Missing required environment variable: {0}")]
    MissingEnv(String),

    #[error("Invalid value for {0}: {1}")]
    InvalidValue(String, String),
}

/// Application configuration
#[derive(Debug, Clone)]
pub struct Config {
    /// PostgreSQL connection URL
    pub database_url: String,

    /// Redis connection URL
    pub redis_url: String,

    /// gRPC server port
    pub grpc_port: u16,

    /// Log level (e.g., "info", "debug", "trace")
    pub log_level: String,

    /// Temporal configuration
    pub temporal: TemporalConfig,
}

/// Temporal workflow engine configuration
#[derive(Debug, Clone)]
pub struct TemporalConfig {
    /// Temporal server host (e.g., "localhost:7233")
    pub host: String,

    /// Temporal namespace
    pub namespace: String,

    /// Whether TLS is enabled for Temporal connection
    pub tls_enabled: bool,

    /// Task queue for courier workflows
    pub task_queue_courier: String,

    /// Task queue for delivery workflows
    pub task_queue_delivery: String,

    /// Worker build ID for versioning
    pub worker_build_id: String,
}

impl Config {
    /// Load configuration from environment variables
    ///
    /// Required env vars:
    /// - DATABASE_URL: PostgreSQL connection string
    /// - REDIS_URL: Redis connection string
    ///
    /// Optional env vars:
    /// - GRPC_PORT: gRPC server port (default: 50051)
    /// - RUST_LOG: Log level (default: "info")
    ///
    /// Temporal env vars:
    /// - TEMPORAL_HOST: Temporal server address (default: localhost:7233)
    /// - TEMPORAL_NAMESPACE: Temporal namespace (default: delivery)
    /// - TEMPORAL_TLS_ENABLED: Whether TLS is enabled (default: false)
    /// - TEMPORAL_TASK_QUEUE_COURIER: Courier task queue (default: COURIER_TASK_QUEUE)
    /// - TEMPORAL_TASK_QUEUE_DELIVERY: Delivery task queue (default: DELIVERY_TASK_QUEUE)
    /// - TEMPORAL_WORKER_BUILD_ID: Worker build ID for versioning (default: delivery-rust-v1)
    ///
    /// Kafka env vars (read by KafkaPublisherConfig/LocationConsumerConfig):
    /// - KAFKA_BROKERS: Kafka bootstrap servers (default: localhost:9092)
    /// - KAFKA_CLIENT_ID: Kafka client ID (default: delivery-service)
    /// - KAFKA_CONSUMER_GROUP: Consumer group for location updates (default: delivery-service)
    /// - KAFKA_MESSAGE_TIMEOUT_MS: Message timeout (default: 5000)
    /// - KAFKA_REQUEST_TIMEOUT_MS: Request timeout (default: 5000)
    /// - KAFKA_LOCATION_TOPIC: Topic for location updates (default: courier.location.updates)
    pub fn from_env() -> Result<Self, ConfigError> {
        // Load .env file if present (ignore errors)
        let _ = dotenvy::dotenv();

        let database_url = env::var("DATABASE_URL")
            .map_err(|_| ConfigError::MissingEnv("DATABASE_URL".to_string()))?;

        let redis_url = env::var("REDIS_URL")
            .map_err(|_| ConfigError::MissingEnv("REDIS_URL".to_string()))?;

        let grpc_port = env::var("GRPC_PORT")
            .unwrap_or_else(|_| "50051".to_string())
            .parse::<u16>()
            .map_err(|e| ConfigError::InvalidValue("GRPC_PORT".to_string(), e.to_string()))?;

        let log_level = env::var("RUST_LOG").unwrap_or_else(|_| "info".to_string());

        // Temporal configuration
        let temporal = TemporalConfig::from_env()?;

        Ok(Self {
            database_url,
            redis_url,
            grpc_port,
            log_level,
            temporal,
        })
    }

    /// Get the gRPC server address
    pub fn grpc_addr(&self) -> String {
        format!("0.0.0.0:{}", self.grpc_port)
    }
}

impl TemporalConfig {
    /// Load Temporal configuration from environment variables
    pub fn from_env() -> Result<Self, ConfigError> {
        let host = env::var("TEMPORAL_HOST")
            .unwrap_or_else(|_| "localhost:7233".to_string());

        let namespace = env::var("TEMPORAL_NAMESPACE")
            .unwrap_or_else(|_| "delivery".to_string());

        let tls_enabled = env::var("TEMPORAL_TLS_ENABLED")
            .unwrap_or_else(|_| "false".to_string())
            .parse::<bool>()
            .map_err(|e| ConfigError::InvalidValue("TEMPORAL_TLS_ENABLED".to_string(), e.to_string()))?;

        let task_queue_courier = env::var("TEMPORAL_TASK_QUEUE_COURIER")
            .unwrap_or_else(|_| "COURIER_TASK_QUEUE".to_string());

        let task_queue_delivery = env::var("TEMPORAL_TASK_QUEUE_DELIVERY")
            .unwrap_or_else(|_| "DELIVERY_TASK_QUEUE".to_string());

        let worker_build_id = env::var("TEMPORAL_WORKER_BUILD_ID")
            .unwrap_or_else(|_| "delivery-rust-v1".to_string());

        Ok(Self {
            host,
            namespace,
            tls_enabled,
            task_queue_courier,
            task_queue_delivery,
            worker_build_id,
        })
    }

    /// Get the Temporal server URL
    pub fn server_url(&self) -> String {
        let scheme = if self.tls_enabled { "https" } else { "http" };
        format!("{}://{}", scheme, self.host)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_grpc_port() {
        // This test requires env vars to be set
        // In practice, use a test helper to set them
    }
}
