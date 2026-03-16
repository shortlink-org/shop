//! Temporal Worker Runner
//!
//! Initializes and runs Temporal workers for the Delivery Service.
//!
//! NOTE: The Temporal Rust SDK is pre-alpha. This implementation provides
//! the basic worker structure and activity registration. Workflow implementations
//! are simplified due to SDK instability.

use std::str::FromStr;
use std::sync::Arc;

use anyhow::anyhow;
use async_trait::async_trait;
use thiserror::Error;
use tracing::info;
use uuid::Uuid;

use temporalio_client::{Client, ClientOptions, Connection, ConnectionOptions, TlsOptions};
use temporalio_common::{
    telemetry::TelemetryOptions,
    worker::{WorkerDeploymentOptions, WorkerTaskTypes},
};
use temporalio_sdk::{
    activities::{activities, ActivityContext, ActivityError},
    Worker, WorkerOptions,
};
use temporalio_sdk_core::{CoreRuntime, RuntimeOptions, Url};

use crate::config::TemporalConfig;
use crate::domain::model::courier::CourierStatus;
use crate::domain::ports::{CourierCache, CourierRepository, PackageRepository};

use super::courier::activities::CourierActivities;
use super::delivery::activities::{DeliveryActivities, DeliveryActivityError};

#[async_trait]
trait CourierTemporalActivityApi: Send + Sync {
    async fn get_free_couriers(&self, zone: &str) -> Result<String, ActivityError>;
    async fn update_courier_status(
        &self,
        courier_id: Uuid,
        status: CourierStatus,
    ) -> Result<String, ActivityError>;
    async fn accept_package(&self, courier_id: Uuid) -> Result<String, ActivityError>;
    async fn complete_courier_delivery(
        &self,
        courier_id: Uuid,
        success: bool,
    ) -> Result<String, ActivityError>;
}

#[async_trait]
impl<R, C> CourierTemporalActivityApi for CourierActivities<R, C>
where
    R: CourierRepository + Send + Sync + 'static,
    C: CourierCache + Send + Sync + 'static,
{
    async fn get_free_couriers(&self, zone: &str) -> Result<String, ActivityError> {
        self.get_free_couriers_in_zone(zone)
            .await
            .map(|couriers| {
                couriers
                    .iter()
                    .map(|c| c.id().to_string())
                    .collect::<Vec<_>>()
                    .join(",")
            })
            .map_err(activity_err)
    }

    async fn update_courier_status(
        &self,
        courier_id: Uuid,
        status: CourierStatus,
    ) -> Result<String, ActivityError> {
        self.update_status(courier_id, status)
            .await
            .map(|_| "ok".to_string())
            .map_err(activity_err)
    }

    async fn accept_package(&self, courier_id: Uuid) -> Result<String, ActivityError> {
        self.accept_package(courier_id)
            .await
            .map(|_| "ok".to_string())
            .map_err(activity_err)
    }

    async fn complete_courier_delivery(
        &self,
        courier_id: Uuid,
        success: bool,
    ) -> Result<String, ActivityError> {
        self.complete_delivery(courier_id, success)
            .await
            .map(|_| "ok".to_string())
            .map_err(activity_err)
    }
}

struct CourierTemporalActivities {
    inner: Arc<dyn CourierTemporalActivityApi>,
}

#[allow(dead_code)]
#[activities]
impl CourierTemporalActivities {
    #[allow(dead_code)]
    #[activity(name = "get_free_couriers")]
    async fn get_free_couriers(
        self: Arc<Self>,
        _ctx: ActivityContext,
        zone: String,
    ) -> Result<String, ActivityError> {
        self.inner.get_free_couriers(&zone).await
    }

    #[allow(dead_code)]
    #[activity(name = "update_courier_status")]
    async fn update_courier_status(
        self: Arc<Self>,
        _ctx: ActivityContext,
        input: String,
    ) -> Result<String, ActivityError> {
        let parts: Vec<&str> = input.split(':').collect();
        if parts.len() != 2 {
            return Err(ActivityError::from(anyhow!(
                "Invalid input format, expected 'id:status'"
            )));
        }

        let courier_id = parse_uuid(parts[0], "courier")?;
        let status = parse_courier_status(parts[1])?;
        self.inner.update_courier_status(courier_id, status).await
    }

    #[allow(dead_code)]
    #[activity(name = "accept_package")]
    async fn accept_package(
        self: Arc<Self>,
        _ctx: ActivityContext,
        courier_id: String,
    ) -> Result<String, ActivityError> {
        let courier_id = parse_uuid(&courier_id, "courier")?;
        self.inner.accept_package(courier_id).await
    }

    #[allow(dead_code)]
    #[activity(name = "complete_courier_delivery")]
    async fn complete_courier_delivery(
        self: Arc<Self>,
        _ctx: ActivityContext,
        input: String,
    ) -> Result<String, ActivityError> {
        let parts: Vec<&str> = input.split(':').collect();
        if parts.len() != 2 {
            return Err(ActivityError::from(anyhow!(
                "Invalid input format, expected 'id:success'"
            )));
        }

        let courier_id = parse_uuid(parts[0], "courier")?;
        let success = parts[1] == "true";
        self.inner
            .complete_courier_delivery(courier_id, success)
            .await
    }
}

#[async_trait]
trait DeliveryTemporalActivityApi: Send + Sync {
    async fn get_dispatch_candidates(&self, zone: &str) -> Result<String, ActivityError>;
    async fn assign_order(&self, courier_id: Uuid, order_id: Uuid)
        -> Result<String, ActivityError>;
    async fn complete_delivery(
        &self,
        courier_id: Uuid,
        order_id: Uuid,
        success: bool,
    ) -> Result<String, ActivityError>;
}

#[async_trait]
impl<R, C, P> DeliveryTemporalActivityApi for DeliveryActivities<R, C, P>
where
    R: CourierRepository + Send + Sync + 'static,
    C: CourierCache + Send + Sync + 'static,
    P: PackageRepository + Send + Sync + 'static,
{
    async fn get_dispatch_candidates(&self, zone: &str) -> Result<String, ActivityError> {
        self.get_dispatch_candidates(zone)
            .await
            .map(|couriers| {
                couriers
                    .iter()
                    .map(|c| c.courier.id().to_string())
                    .collect::<Vec<_>>()
                    .join(",")
            })
            .map_err(delivery_activity_err)
    }

    async fn assign_order(
        &self,
        courier_id: Uuid,
        order_id: Uuid,
    ) -> Result<String, ActivityError> {
        self.assign_order(courier_id, order_id)
            .await
            .map(|_| "ok".to_string())
            .map_err(delivery_activity_err)
    }

    async fn complete_delivery(
        &self,
        courier_id: Uuid,
        order_id: Uuid,
        success: bool,
    ) -> Result<String, ActivityError> {
        self.complete_delivery(courier_id, order_id, success)
            .await
            .map(|_| "ok".to_string())
            .map_err(delivery_activity_err)
    }
}

struct DeliveryTemporalActivities {
    inner: Arc<dyn DeliveryTemporalActivityApi>,
}

#[allow(dead_code)]
#[activities]
impl DeliveryTemporalActivities {
    #[allow(dead_code)]
    #[activity(name = "get_dispatch_candidates")]
    async fn get_dispatch_candidates(
        self: Arc<Self>,
        _ctx: ActivityContext,
        zone: String,
    ) -> Result<String, ActivityError> {
        self.inner.get_dispatch_candidates(&zone).await
    }

    #[allow(dead_code)]
    #[activity(name = "assign_order")]
    async fn assign_order(
        self: Arc<Self>,
        _ctx: ActivityContext,
        input: String,
    ) -> Result<String, ActivityError> {
        let parts: Vec<&str> = input.split(':').collect();
        if parts.len() != 2 {
            return Err(ActivityError::from(anyhow!(
                "Invalid input format, expected 'courier_id:order_id'"
            )));
        }

        let courier_id = parse_uuid(parts[0], "courier")?;
        let order_id = parse_uuid(parts[1], "order")?;
        self.inner.assign_order(courier_id, order_id).await
    }

    #[allow(dead_code)]
    #[activity(name = "complete_delivery")]
    async fn complete_delivery(
        self: Arc<Self>,
        _ctx: ActivityContext,
        input: String,
    ) -> Result<String, ActivityError> {
        let parts: Vec<&str> = input.split(':').collect();
        if parts.len() != 3 {
            return Err(ActivityError::from(anyhow!(
                "Invalid input format, expected 'courier_id:order_id:success'"
            )));
        }

        let courier_id = parse_uuid(parts[0], "courier")?;
        let order_id = parse_uuid(parts[1], "order")?;
        let success = parts[2] == "true";

        self.inner
            .complete_delivery(courier_id, order_id, success)
            .await
    }
}

/// Errors from worker initialization and runtime
#[derive(Debug, Error)]
pub enum WorkerError {
    #[error("Failed to parse Temporal URL: {0}")]
    UrlParseError(String),

    #[error("Failed to create Temporal runtime: {0}")]
    RuntimeError(String),

    #[error("Failed to connect to Temporal: {0}")]
    ConnectionError(String),

    #[error("Failed to create Temporal client: {0}")]
    ClientError(String),

    #[error("Failed to create worker: {0}")]
    WorkerCreationError(String),

    #[error("Worker execution error: {0}")]
    ExecutionError(String),
}

/// Temporal worker runner
///
/// Manages the lifecycle of Temporal workers for courier and delivery workflows.
pub struct TemporalWorkerRunner {
    config: TemporalConfig,
    runtime: CoreRuntime,
}

impl TemporalWorkerRunner {
    /// Create a new worker runner
    pub fn new(config: TemporalConfig) -> Result<Self, WorkerError> {
        let telemetry_options = TelemetryOptions::builder().build();
        let runtime_options = RuntimeOptions::builder()
            .telemetry_options(telemetry_options)
            .build()
            .map_err(|e| WorkerError::RuntimeError(e.to_string()))?;

        let runtime = CoreRuntime::new_assume_tokio(runtime_options)
            .map_err(|e| WorkerError::RuntimeError(e.to_string()))?;

        Ok(Self { config, runtime })
    }

    /// Run the courier worker
    ///
    /// This worker handles courier lifecycle workflows:
    /// - Courier registration
    /// - Status updates (online/offline)
    /// - Package acceptance and delivery completion
    pub async fn run_courier_worker<R, C>(
        &self,
        courier_activities: Arc<CourierActivities<R, C>>,
    ) -> Result<(), WorkerError>
    where
        R: CourierRepository + Send + Sync + 'static,
        C: CourierCache + Send + Sync + 'static,
    {
        info!(
            "Starting courier worker on task queue: {}",
            self.config.task_queue_courier
        );

        let client = self.connect_client().await?;
        let worker_options = self.activity_worker_options(&self.config.task_queue_courier);
        let mut worker = Worker::new(&self.runtime, client, worker_options)
            .map_err(|e| WorkerError::WorkerCreationError(e.to_string()))?;

        worker.register_activities(CourierTemporalActivities {
            inner: courier_activities,
        });

        info!("Courier worker started, polling for tasks...");

        worker
            .run()
            .await
            .map_err(|e| WorkerError::ExecutionError(e.to_string()))
    }

    /// Run the delivery worker
    ///
    /// This worker handles delivery workflows:
    /// - Order assignment to couriers
    /// - Delivery tracking and completion
    pub async fn run_delivery_worker<R, C, P>(
        &self,
        delivery_activities: Arc<DeliveryActivities<R, C, P>>,
    ) -> Result<(), WorkerError>
    where
        R: CourierRepository + Send + Sync + 'static,
        C: CourierCache + Send + Sync + 'static,
        P: PackageRepository + Send + Sync + 'static,
    {
        info!(
            "Starting delivery worker on task queue: {}",
            self.config.task_queue_delivery
        );

        let client = self.connect_client().await?;
        let worker_options = self.activity_worker_options(&self.config.task_queue_delivery);
        let mut worker = Worker::new(&self.runtime, client, worker_options)
            .map_err(|e| WorkerError::WorkerCreationError(e.to_string()))?;

        worker.register_activities(DeliveryTemporalActivities {
            inner: delivery_activities,
        });

        info!("Delivery worker started, polling for tasks...");

        worker
            .run()
            .await
            .map_err(|e| WorkerError::ExecutionError(e.to_string()))
    }

    async fn connect_client(&self) -> Result<Client, WorkerError> {
        let url = Url::from_str(&self.config.server_url())
            .map_err(|e| WorkerError::UrlParseError(e.to_string()))?;

        let connection_options = if self.config.tls_enabled {
            ConnectionOptions::new(url)
                .tls_options(TlsOptions::default())
                .build()
        } else {
            ConnectionOptions::new(url).build()
        };

        let connection = Connection::connect(connection_options)
            .await
            .map_err(|e| WorkerError::ConnectionError(e.to_string()))?;

        Client::new(
            connection,
            ClientOptions::new(self.config.namespace.clone()).build(),
        )
        .map_err(|e| WorkerError::ClientError(e.to_string()))
    }

    fn activity_worker_options(&self, task_queue: &str) -> WorkerOptions {
        WorkerOptions::new(task_queue.to_string())
            .task_types(WorkerTaskTypes::activity_only())
            .deployment_options(WorkerDeploymentOptions::from_build_id(
                self.config.worker_build_id.clone(),
            ))
            .build()
    }
}

fn parse_uuid(value: &str, entity: &str) -> Result<Uuid, ActivityError> {
    Uuid::parse_str(value).map_err(|e| ActivityError::from(anyhow!("Invalid {entity} ID: {e}")))
}

fn parse_courier_status(value: &str) -> Result<CourierStatus, ActivityError> {
    match value {
        "free" => Ok(CourierStatus::Free),
        "busy" => Ok(CourierStatus::Busy),
        "unavailable" => Ok(CourierStatus::Unavailable),
        _ => Err(ActivityError::from(anyhow!("Invalid status: {value}"))),
    }
}

fn activity_err<E>(error: E) -> ActivityError
where
    E: std::fmt::Display,
{
    ActivityError::from(anyhow!("{error}"))
}

fn delivery_activity_err(error: DeliveryActivityError) -> ActivityError {
    match error {
        DeliveryActivityError::CourierUnavailable(courier_id) => {
            ActivityError::NonRetryable(Box::new(DeliveryActivityError::CourierUnavailable(
                courier_id,
            )))
        }
        other => activity_err(other),
    }
}

// =============================================================================
// Workflow Registration Notes
//
// The Temporal Rust SDK is pre-alpha and the workflow API is unstable.
// Workflow definitions are in courier/workflow.rs and delivery/workflow.rs.
//
// To register workflows, use the typed workflow API exposed by the current SDK.
//
// Workflows can:
// - Call activities via typed activity handles
// - Use timers
// - Handle signals
// - Query state via query handlers
//
// See the SDK documentation for the latest API.
// =============================================================================
