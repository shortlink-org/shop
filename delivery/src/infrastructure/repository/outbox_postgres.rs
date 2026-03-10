//! PostgreSQL Outbox Repository
//!
//! Handles transactional outbox inserts and forwarder polling state.

use std::time::Duration;

use chrono::{Duration as ChronoDuration, Utc};
use sea_orm::{
    ActiveModelTrait, ActiveValue::Set, DatabaseConnection, DatabaseTransaction, DbBackend,
    EntityTrait, IntoActiveModel, Statement,
};
use uuid::Uuid;

use crate::domain::ports::{DomainEvent, RepositoryError};
use crate::infrastructure::messaging::kafka_publisher::{
    KafkaEventPublisher, KafkaOutboundRecord, KafkaPayloadEncoding,
};
use crate::infrastructure::repository::entities::outbox_message::{self, Entity as OutboxEntity};

/// PostgreSQL implementation of the delivery outbox.
pub struct OutboxPostgresRepository {
    db: DatabaseConnection,
}

impl OutboxPostgresRepository {
    /// Create a new outbox repository instance.
    pub fn new(db: DatabaseConnection) -> Self {
        Self { db }
    }

    /// Insert one or more domain events into outbox within an existing transaction.
    pub async fn insert_events_in_tx(
        tx: &DatabaseTransaction,
        events: &[DomainEvent],
    ) -> Result<(), RepositoryError> {
        if events.is_empty() {
            return Ok(());
        }

        let now = Utc::now();
        let mut models = Vec::new();

        for event in events {
            let records = KafkaEventPublisher::encode_event(event)
                .map_err(|e| RepositoryError::SerializationError(e.to_string()))?;

            for record in records {
                models.push(Self::active_model_from_record(record, now)?);
            }
        }

        if models.is_empty() {
            return Ok(());
        }

        OutboxEntity::insert_many(models)
            .exec(tx)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        Ok(())
    }

    fn active_model_from_record(
        record: KafkaOutboundRecord,
        now: chrono::DateTime<Utc>,
    ) -> Result<outbox_message::ActiveModel, RepositoryError> {
        let headers = serde_json::to_string(&record.headers)
            .map_err(|e| RepositoryError::SerializationError(e.to_string()))?;

        Ok(outbox_message::ActiveModel {
            id: Set(Uuid::new_v4()),
            topic: Set(record.topic),
            message_key: Set(record.key),
            payload: Set(record.payload),
            headers: Set(headers),
            payload_encoding: Set(record.payload_encoding.as_str().to_string()),
            event_type: Set(record.event_type),
            aggregate_id: Set(record.aggregate_id),
            available_at: Set(now),
            locked_at: Set(None),
            locked_by: Set(None),
            processed_at: Set(None),
            attempts: Set(0),
            last_error: Set(None),
            created_at: Set(now),
        })
    }

    /// Claim a batch of pending outbox messages for publishing.
    pub async fn claim_batch(
        &self,
        worker_id: &str,
        batch_size: u64,
        lock_timeout: Duration,
    ) -> Result<Vec<outbox_message::Model>, RepositoryError> {
        let lock_expired_before = Utc::now()
            - ChronoDuration::from_std(lock_timeout)
                .unwrap_or_else(|_| ChronoDuration::seconds(30));

        let statement = Statement::from_sql_and_values(
            DbBackend::Postgres,
            r#"
            WITH candidates AS (
                SELECT id
                FROM delivery.outbox_messages
                WHERE processed_at IS NULL
                  AND available_at <= NOW()
                  AND (locked_at IS NULL OR locked_at < $1)
                ORDER BY created_at
                LIMIT $2
                FOR UPDATE SKIP LOCKED
            )
            UPDATE delivery.outbox_messages AS o
            SET locked_at = NOW(),
                locked_by = $3
            FROM candidates
            WHERE o.id = candidates.id
            RETURNING
                o.id,
                o.topic,
                o.message_key,
                o.payload,
                o.headers,
                o.payload_encoding,
                o.event_type,
                o.aggregate_id,
                o.available_at,
                o.locked_at,
                o.locked_by,
                o.processed_at,
                o.attempts,
                o.last_error,
                o.created_at
            "#,
            [
                lock_expired_before.into(),
                (batch_size as i64).into(),
                worker_id.to_string().into(),
            ],
        );

        OutboxEntity::find()
            .from_raw_sql(statement)
            .all(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))
    }

    /// Mark an outbox row as successfully published.
    pub async fn mark_published(&self, id: Uuid) -> Result<(), RepositoryError> {
        let Some(model) = OutboxEntity::find_by_id(id)
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?
        else {
            return Ok(());
        };

        let attempts = model.attempts;
        let mut active_model = model.into_active_model();
        active_model.processed_at = Set(Some(Utc::now()));
        active_model.locked_at = Set(None);
        active_model.locked_by = Set(None);
        active_model.last_error = Set(None);
        active_model.attempts = Set(attempts + 1);
        active_model
            .update(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        Ok(())
    }

    /// Mark an outbox row as failed and schedule it for retry.
    pub async fn mark_failed(
        &self,
        id: Uuid,
        error: &str,
        retry_delay: Duration,
    ) -> Result<(), RepositoryError> {
        let Some(model) = OutboxEntity::find_by_id(id)
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?
        else {
            return Ok(());
        };

        let attempts = model.attempts;
        let available_at = Utc::now()
            + ChronoDuration::from_std(retry_delay).unwrap_or_else(|_| ChronoDuration::seconds(5));

        let mut active_model = model.into_active_model();
        active_model.available_at = Set(available_at);
        active_model.locked_at = Set(None);
        active_model.locked_by = Set(None);
        active_model.last_error = Set(Some(error.to_string()));
        active_model.attempts = Set(attempts + 1);
        active_model
            .update(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        Ok(())
    }

    /// Convert an outbox row back to a Kafka record for forwarding.
    pub(crate) fn to_kafka_record(
        model: &outbox_message::Model,
    ) -> Result<KafkaOutboundRecord, RepositoryError> {
        let headers: Vec<(String, String)> = serde_json::from_str(&model.headers)
            .map_err(|e| RepositoryError::SerializationError(e.to_string()))?;

        let payload_encoding = match model.payload_encoding.as_str() {
            "json" => KafkaPayloadEncoding::Json,
            _ => KafkaPayloadEncoding::Proto,
        };

        Ok(KafkaOutboundRecord {
            topic: model.topic.clone(),
            key: model.message_key.clone(),
            payload: model.payload.clone(),
            headers,
            payload_encoding,
            event_type: model.event_type.clone(),
            aggregate_id: model.aggregate_id.clone(),
        })
    }
}
