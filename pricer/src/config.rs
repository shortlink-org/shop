//! Configuration
//!
//! Loads from environment variables. Optional: .env file via dotenvy.

use std::env;

use thiserror::Error;

#[derive(Debug, Error)]
pub enum ConfigError {
    #[error("Missing required environment variable: {0}")]
    MissingEnv(String),

    #[error("Invalid value for {0}: {1}")]
    InvalidValue(String, String),
}

#[derive(Debug, Clone)]
pub struct Config {
    /// Log level (e.g. "info", "debug")
    pub log_level: String,

    /// Path to OPA discount policies (e.g. "policies/discounts/")
    pub policies_discounts: String,

    /// Path to OPA tax policies (e.g. "policies/taxes/")
    pub policies_taxes: String,

    /// gRPC server port (for Phase 4)
    pub grpc_port: u16,
}

impl Config {
    /// Load configuration from environment.
    ///
    /// Optional env vars:
    /// - RUST_LOG: log level (default: "info")
    /// - PRICER_POLICIES_DISCOUNTS: path to discount policies (default: "policies/discounts")
    /// - PRICER_POLICIES_TAXES: path to tax policies (default: "policies/taxes")
    /// - PRICER_GRPC_PORT: gRPC port (default: 50052)
    pub fn from_env() -> Result<Self, ConfigError> {
        let _ = dotenvy::dotenv();

        let log_level = env::var("RUST_LOG").unwrap_or_else(|_| "info".to_string());
        let policies_discounts = env::var("PRICER_POLICIES_DISCOUNTS")
            .unwrap_or_else(|_| "policies/discounts".to_string());
        let policies_taxes = env::var("PRICER_POLICIES_TAXES")
            .unwrap_or_else(|_| "policies/taxes".to_string());
        let grpc_port: u16 = env::var("PRICER_GRPC_PORT")
            .unwrap_or_else(|_| "50052".to_string())
            .parse()
            .map_err(|_| ConfigError::InvalidValue("PRICER_GRPC_PORT".into(), "must be u16".into()))?;

        Ok(Config {
            log_level,
            policies_discounts,
            policies_taxes,
            grpc_port,
        })
    }
}
