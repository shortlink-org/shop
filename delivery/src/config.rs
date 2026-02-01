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

        Ok(Self {
            database_url,
            redis_url,
            grpc_port,
            log_level,
        })
    }

    /// Get the gRPC server address
    pub fn grpc_addr(&self) -> String {
        format!("0.0.0.0:{}", self.grpc_port)
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
