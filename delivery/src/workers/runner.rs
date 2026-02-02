//! Temporal Worker Runner
//!
//! Initializes and runs Temporal workers for the Delivery Service.
//!
//! NOTE: The Temporal Rust SDK is pre-alpha. This implementation provides
//! the basic worker structure and activity registration. Workflow implementations
//! are simplified due to SDK instability.

use std::str::FromStr;
use std::sync::Arc;

use thiserror::Error;
use tracing::info;

use temporalio_common::{
    telemetry::TelemetryOptions,
    worker::{WorkerConfig, WorkerTaskTypes, WorkerVersioningStrategy},
};
use temporalio_sdk::{sdk_client_options, ActContext, Worker};
use temporalio_sdk_core::{init_worker, CoreRuntime, RuntimeOptions, Url};

use crate::config::TemporalConfig;
use crate::domain::model::courier::CourierStatus;

use super::courier::activities::CourierActivities;
use super::delivery::activities::DeliveryActivities;

/// Errors from worker initialization and runtime
#[derive(Debug, Error)]
pub enum WorkerError {
    #[error("Failed to parse Temporal URL: {0}")]
    UrlParseError(String),

    #[error("Failed to create Temporal runtime: {0}")]
    RuntimeError(String),

    #[error("Failed to connect to Temporal: {0}")]
    ConnectionError(String),

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
        // Initialize telemetry
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
        R: crate::domain::ports::CourierRepository + Send + Sync + 'static,
        C: crate::domain::ports::CourierCache + Send + Sync + 'static,
    {
        info!(
            "Starting courier worker on task queue: {}",
            self.config.task_queue_courier
        );

        let url = Url::from_str(&self.config.server_url())
            .map_err(|e| WorkerError::UrlParseError(e.to_string()))?;

        let client_options = sdk_client_options(url).build();
        let client = client_options
            .connect(&self.config.namespace, None)
            .await
            .map_err(|e| WorkerError::ConnectionError(e.to_string()))?;

        let worker_config = WorkerConfig::builder()
            .namespace(&self.config.namespace)
            .task_queue(&self.config.task_queue_courier)
            .task_types(WorkerTaskTypes::activity_only())
            .versioning_strategy(WorkerVersioningStrategy::None {
                build_id: self.config.worker_build_id.clone(),
            })
            .build()
            .map_err(|e| WorkerError::WorkerCreationError(e.to_string()))?;

        let core_worker = init_worker(&self.runtime, worker_config, client)
            .map_err(|e| WorkerError::WorkerCreationError(e.to_string()))?;

        let mut worker =
            Worker::new_from_core(Arc::new(core_worker), &self.config.task_queue_courier);

        // Register courier activities
        self.register_courier_activities(&mut worker, courier_activities);

        info!("Courier worker started, polling for tasks...");

        worker
            .run()
            .await
            .map_err(|e| WorkerError::ExecutionError(e.to_string()))?;

        Ok(())
    }

    /// Run the delivery worker
    ///
    /// This worker handles delivery workflows:
    /// - Order assignment to couriers
    /// - Delivery tracking and completion
    pub async fn run_delivery_worker<R, C>(
        &self,
        delivery_activities: Arc<DeliveryActivities<R, C>>,
    ) -> Result<(), WorkerError>
    where
        R: crate::domain::ports::CourierRepository + Send + Sync + 'static,
        C: crate::domain::ports::CourierCache + Send + Sync + 'static,
    {
        info!(
            "Starting delivery worker on task queue: {}",
            self.config.task_queue_delivery
        );

        let url = Url::from_str(&self.config.server_url())
            .map_err(|e| WorkerError::UrlParseError(e.to_string()))?;

        let client_options = sdk_client_options(url).build();
        let client = client_options
            .connect(&self.config.namespace, None)
            .await
            .map_err(|e| WorkerError::ConnectionError(e.to_string()))?;

        let worker_config = WorkerConfig::builder()
            .namespace(&self.config.namespace)
            .task_queue(&self.config.task_queue_delivery)
            .task_types(WorkerTaskTypes::activity_only())
            .versioning_strategy(WorkerVersioningStrategy::None {
                build_id: self.config.worker_build_id.clone(),
            })
            .build()
            .map_err(|e| WorkerError::WorkerCreationError(e.to_string()))?;

        let core_worker = init_worker(&self.runtime, worker_config, client)
            .map_err(|e| WorkerError::WorkerCreationError(e.to_string()))?;

        let mut worker =
            Worker::new_from_core(Arc::new(core_worker), &self.config.task_queue_delivery);

        // Register delivery activities
        self.register_delivery_activities(&mut worker, delivery_activities);

        info!("Delivery worker started, polling for tasks...");

        worker
            .run()
            .await
            .map_err(|e| WorkerError::ExecutionError(e.to_string()))?;

        Ok(())
    }

    /// Register courier activities with the worker
    fn register_courier_activities<R, C>(
        &self,
        worker: &mut Worker,
        activities: Arc<CourierActivities<R, C>>,
    ) where
        R: crate::domain::ports::CourierRepository + Send + Sync + 'static,
        C: crate::domain::ports::CourierCache + Send + Sync + 'static,
    {
        use temporalio_sdk::ActivityError;

        // Register get_free_couriers activity
        let acts = activities.clone();
        worker.register_activity(
            "get_free_couriers",
            move |_ctx: ActContext, zone: String| {
                let acts = acts.clone();
                async move {
                    acts.get_free_couriers_in_zone(&zone)
                        .await
                        .map(|couriers| {
                            couriers
                                .iter()
                                .map(|c| c.id().to_string())
                                .collect::<Vec<_>>()
                                .join(",")
                        })
                        .map_err(|e| ActivityError::from(anyhow::anyhow!("{}", e)))
                }
            },
        );

        // Register update_status activity
        let acts = activities.clone();
        worker.register_activity(
            "update_courier_status",
            move |_ctx: ActContext, input: String| {
                let acts = acts.clone();
                async move {
                    // Parse input as "courier_id:status"
                    let parts: Vec<&str> = input.split(':').collect();
                    if parts.len() != 2 {
                        return Err(ActivityError::from(anyhow::anyhow!(
                            "Invalid input format, expected 'id:status'"
                        )));
                    }

                    let courier_id = uuid::Uuid::parse_str(parts[0]).map_err(|e| {
                        ActivityError::from(anyhow::anyhow!("Invalid courier ID: {}", e))
                    })?;

                    let status = match parts[1] {
                        "free" => CourierStatus::Free,
                        "busy" => CourierStatus::Busy,
                        "unavailable" => CourierStatus::Unavailable,
                        _ => {
                            return Err(ActivityError::from(anyhow::anyhow!(
                                "Invalid status: {}",
                                parts[1]
                            )))
                        }
                    };

                    acts.update_status(courier_id, status)
                        .await
                        .map(|_| "ok".to_string())
                        .map_err(|e| ActivityError::from(anyhow::anyhow!("{}", e)))
                }
            },
        );

        // Register accept_package activity
        let acts = activities.clone();
        worker.register_activity(
            "accept_package",
            move |_ctx: ActContext, courier_id: String| {
                let acts = acts.clone();
                async move {
                    let id = uuid::Uuid::parse_str(&courier_id).map_err(|e| {
                        ActivityError::from(anyhow::anyhow!("Invalid courier ID: {}", e))
                    })?;
                    acts.accept_package(id)
                        .await
                        .map(|_| "ok".to_string())
                        .map_err(|e| ActivityError::from(anyhow::anyhow!("{}", e)))
                }
            },
        );

        // Register complete_delivery activity
        let acts = activities.clone();
        worker.register_activity(
            "complete_courier_delivery",
            move |_ctx: ActContext, input: String| {
                let acts = acts.clone();
                async move {
                    // Parse input as "courier_id:success"
                    let parts: Vec<&str> = input.split(':').collect();
                    if parts.len() != 2 {
                        return Err(ActivityError::from(anyhow::anyhow!(
                            "Invalid input format, expected 'id:success'"
                        )));
                    }

                    let courier_id = uuid::Uuid::parse_str(parts[0]).map_err(|e| {
                        ActivityError::from(anyhow::anyhow!("Invalid courier ID: {}", e))
                    })?;

                    let success = parts[1] == "true";

                    acts.complete_delivery(courier_id, success)
                        .await
                        .map(|_| "ok".to_string())
                        .map_err(|e| ActivityError::from(anyhow::anyhow!("{}", e)))
                }
            },
        );

        info!("Registered courier activities");
    }

    /// Register delivery activities with the worker
    fn register_delivery_activities<R, C>(
        &self,
        worker: &mut Worker,
        activities: Arc<DeliveryActivities<R, C>>,
    ) where
        R: crate::domain::ports::CourierRepository + Send + Sync + 'static,
        C: crate::domain::ports::CourierCache + Send + Sync + 'static,
    {
        use temporalio_sdk::ActivityError;

        // Register get_free_couriers_for_dispatch activity
        let acts = activities.clone();
        worker.register_activity(
            "get_free_couriers_for_dispatch",
            move |_ctx: ActContext, zone: String| {
                let acts = acts.clone();
                async move {
                    acts.get_free_couriers_for_dispatch(&zone)
                        .await
                        .map(|couriers| {
                            // Return courier IDs as comma-separated string
                            couriers
                                .iter()
                                .map(|c| c.id.clone())
                                .collect::<Vec<_>>()
                                .join(",")
                        })
                        .map_err(|e| ActivityError::from(anyhow::anyhow!("{}", e)))
                }
            },
        );

        // Register assign_order activity
        let acts = activities.clone();
        worker.register_activity(
            "assign_order",
            move |_ctx: ActContext, input: String| {
                let acts = acts.clone();
                async move {
                    // Parse input as "courier_id:order_id"
                    let parts: Vec<&str> = input.split(':').collect();
                    if parts.len() != 2 {
                        return Err(ActivityError::from(anyhow::anyhow!(
                            "Invalid input format, expected 'courier_id:order_id'"
                        )));
                    }

                    let courier_id = uuid::Uuid::parse_str(parts[0]).map_err(|e| {
                        ActivityError::from(anyhow::anyhow!("Invalid courier ID: {}", e))
                    })?;
                    let order_id = uuid::Uuid::parse_str(parts[1]).map_err(|e| {
                        ActivityError::from(anyhow::anyhow!("Invalid order ID: {}", e))
                    })?;

                    acts.assign_order(courier_id, order_id)
                        .await
                        .map(|_| "ok".to_string())
                        .map_err(|e| ActivityError::from(anyhow::anyhow!("{}", e)))
                }
            },
        );

        // Register complete_delivery activity
        let acts = activities.clone();
        worker.register_activity(
            "complete_delivery",
            move |_ctx: ActContext, input: String| {
                let acts = acts.clone();
                async move {
                    // Parse input as "courier_id:order_id:success"
                    let parts: Vec<&str> = input.split(':').collect();
                    if parts.len() != 3 {
                        return Err(ActivityError::from(anyhow::anyhow!(
                            "Invalid input format, expected 'courier_id:order_id:success'"
                        )));
                    }

                    let courier_id = uuid::Uuid::parse_str(parts[0]).map_err(|e| {
                        ActivityError::from(anyhow::anyhow!("Invalid courier ID: {}", e))
                    })?;
                    let order_id = uuid::Uuid::parse_str(parts[1]).map_err(|e| {
                        ActivityError::from(anyhow::anyhow!("Invalid order ID: {}", e))
                    })?;
                    let success = parts[2] == "true";

                    acts.complete_delivery(courier_id, order_id, success)
                        .await
                        .map(|_| "ok".to_string())
                        .map_err(|e| ActivityError::from(anyhow::anyhow!("{}", e)))
                }
            },
        );

        info!("Registered delivery activities");
    }
}

// =============================================================================
// Workflow Registration Notes
//
// The Temporal Rust SDK is pre-alpha and the workflow API is unstable.
// Workflow definitions are in courier/workflow.rs and delivery/workflow.rs.
//
// To register workflows, use:
// ```rust
// worker.register_wf("workflow_name", WorkflowFunction::new(|ctx: WfContext| async move {
//     // Workflow logic here
//     Ok(WfExitValue::Normal(result))
// }));
// ```
//
// Workflows can:
// - Call activities: ctx.activity(ActivityOptions { ... }).await
// - Use timers: ctx.timer(duration).await
// - Handle signals: ctx.make_signal_channel("name")
// - Query state: via query handlers
//
// See the SDK documentation for the latest API.
// =============================================================================
