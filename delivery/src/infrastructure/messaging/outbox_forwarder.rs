//! Delivery Outbox Forwarder
//!
//! Polls transactional outbox rows and forwards them to Kafka.

use std::time::Duration;

use tokio::sync::broadcast;
use tracing::{error, info};

use crate::infrastructure::messaging::kafka_publisher::KafkaEventPublisher;
use crate::infrastructure::repository::OutboxPostgresRepository;

/// Configuration for outbox forwarder polling and retries.
#[derive(Debug, Clone)]
pub struct OutboxForwarderConfig {
    pub worker_id: String,
    pub poll_interval_ms: u64,
    pub batch_size: u64,
    pub retry_delay_ms: u64,
    pub lock_timeout_ms: u64,
}

impl Default for OutboxForwarderConfig {
    fn default() -> Self {
        Self {
            worker_id: format!("delivery-{}", std::process::id()),
            poll_interval_ms: 1000,
            batch_size: 100,
            retry_delay_ms: 5000,
            lock_timeout_ms: 30000,
        }
    }
}

impl OutboxForwarderConfig {
    pub fn from_env() -> Self {
        Self {
            worker_id: std::env::var("OUTBOX_WORKER_ID")
                .unwrap_or_else(|_| format!("delivery-{}", std::process::id())),
            poll_interval_ms: std::env::var("OUTBOX_POLL_INTERVAL_MS")
                .ok()
                .and_then(|s| s.parse().ok())
                .unwrap_or(1000),
            batch_size: std::env::var("OUTBOX_BATCH_SIZE")
                .ok()
                .and_then(|s| s.parse().ok())
                .unwrap_or(100),
            retry_delay_ms: std::env::var("OUTBOX_RETRY_DELAY_MS")
                .ok()
                .and_then(|s| s.parse().ok())
                .unwrap_or(5000),
            lock_timeout_ms: std::env::var("OUTBOX_LOCK_TIMEOUT_MS")
                .ok()
                .and_then(|s| s.parse().ok())
                .unwrap_or(30000),
        }
    }
}

/// Background outbox forwarder.
pub struct OutboxForwarder {
    config: OutboxForwarderConfig,
    outbox_repo: std::sync::Arc<OutboxPostgresRepository>,
    publisher: std::sync::Arc<KafkaEventPublisher>,
    shutdown_rx: broadcast::Receiver<()>,
}

impl OutboxForwarder {
    pub fn new(
        config: OutboxForwarderConfig,
        outbox_repo: std::sync::Arc<OutboxPostgresRepository>,
        publisher: std::sync::Arc<KafkaEventPublisher>,
        shutdown_rx: broadcast::Receiver<()>,
    ) -> Self {
        Self {
            config,
            outbox_repo,
            publisher,
            shutdown_rx,
        }
    }

    pub async fn run(mut self) {
        let poll_interval = Duration::from_millis(self.config.poll_interval_ms);
        let retry_delay = Duration::from_millis(self.config.retry_delay_ms);
        let lock_timeout = Duration::from_millis(self.config.lock_timeout_ms);

        info!(
            worker_id = %self.config.worker_id,
            batch_size = self.config.batch_size,
            "Starting delivery outbox forwarder"
        );

        loop {
            tokio::select! {
                _ = self.shutdown_rx.recv() => {
                    info!("Outbox forwarder received shutdown signal");
                    break;
                }
                _ = tokio::time::sleep(poll_interval) => {
                    let batch = match self
                        .outbox_repo
                        .claim_batch(&self.config.worker_id, self.config.batch_size, lock_timeout)
                        .await
                    {
                        Ok(batch) => batch,
                        Err(err) => {
                            error!(error = %err, "Failed to claim delivery outbox batch");
                            continue;
                        }
                    };

                    for message in batch {
                        let record = match OutboxPostgresRepository::to_kafka_record(&message) {
                            Ok(record) => record,
                            Err(err) => {
                                error!(message_id = %message.id, error = %err, "Failed to decode outbox message");
                                if let Err(mark_err) = self
                                    .outbox_repo
                                    .mark_failed(message.id, &err.to_string(), retry_delay)
                                    .await
                                {
                                    error!(message_id = %message.id, error = %mark_err, "Failed to mark outbox message as failed");
                                }
                                continue;
                            }
                        };

                        match self.publisher.publish_record(&record).await {
                            Ok(()) => {
                                if let Err(err) = self.outbox_repo.mark_published(message.id).await {
                                    error!(message_id = %message.id, error = %err, "Failed to mark outbox message as published");
                                }
                            }
                            Err(err) => {
                                error!(
                                    message_id = %message.id,
                                    topic = %message.topic,
                                    error = %err,
                                    "Failed to forward outbox message to Kafka"
                                );
                                if let Err(mark_err) = self
                                    .outbox_repo
                                    .mark_failed(message.id, &err.to_string(), retry_delay)
                                    .await
                                {
                                    error!(message_id = %message.id, error = %mark_err, "Failed to mark outbox message as failed");
                                }
                            }
                        }
                    }
                }
            }
        }

        info!("Delivery outbox forwarder stopped");
    }
}
