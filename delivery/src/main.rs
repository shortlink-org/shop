//! Delivery Service
//!
//! gRPC server for delivery management operations with Temporal workflow integration.

use std::sync::Arc;

use tonic::transport::Server;
use tonic_health::server::health_reporter;
use tracing::{error, info, warn};
use tracing_subscriber::EnvFilter;

use delivery::config::Config;
use delivery::di::AppState;
use delivery::infrastructure::rpc::server::DeliveryServiceImpl;
use delivery::infrastructure::rpc::DeliveryServiceServer;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Load configuration
    let config = Config::from_env().map_err(|e| {
        eprintln!("Configuration error: {}", e);
        e
    })?;

    // Initialize logging
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::try_from_default_env().unwrap_or_else(|_| {
            EnvFilter::new(&config.log_level)
        }))
        .init();

    info!("Starting Delivery Service...");

    // Initialize application state
    let state = Arc::new(AppState::new(&config).await.map_err(|e| {
        error!(error = %e, "Failed to initialize application state");
        e
    })?);

    // Start background consumers (location updates from Kafka)
    if let Err(e) = state.start_consumers().await {
        warn!(error = %e, "Failed to start Kafka consumers (continuing without real-time location updates)");
    }

    // Start Temporal workers (courier and delivery workflows)
    if let Err(e) = state.start_temporal_workers(&config.temporal).await {
        warn!(error = %e, "Failed to start Temporal workers (continuing without workflow orchestration)");
        warn!("Ensure Temporal server is running at: {}", config.temporal.server_url());
    }

    // Create gRPC health service
    let (health_reporter, health_service) = health_reporter();
    // Set the service as serving
    health_reporter
        .set_serving::<DeliveryServiceServer<DeliveryServiceImpl>>()
        .await;

    // Create gRPC service
    let delivery_service = DeliveryServiceImpl::new(state.clone());

    // Start gRPC server
    let addr = config.grpc_addr().parse()?;
    info!(address = %addr, "gRPC server starting");
    info!(
        temporal_host = %config.temporal.host,
        temporal_namespace = %config.temporal.namespace,
        "Temporal configuration"
    );

    // Handle graceful shutdown
    let state_for_shutdown = state.clone();
    tokio::spawn(async move {
        if let Err(e) = tokio::signal::ctrl_c().await {
            error!(error = %e, "Failed to listen for ctrl-c signal");
            return;
        }
        info!("Received shutdown signal");
        state_for_shutdown.shutdown();
    });

    Server::builder()
        .add_service(health_service)
        .add_service(DeliveryServiceServer::new(delivery_service))
        .serve(addr)
        .await?;

    Ok(())
}

