//! Outbox Message Entity for Sea-ORM
//!
//! Database entity for the delivery outbox table.

use chrono::{DateTime, Utc};
use sea_orm::entity::prelude::*;
use uuid::Uuid;

#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(schema_name = "delivery", table_name = "outbox_messages")]
pub struct Model {
    #[sea_orm(primary_key, auto_increment = false)]
    pub id: Uuid,
    pub topic: String,
    pub message_key: String,
    pub payload: Vec<u8>,
    pub headers: String,
    pub payload_encoding: String,
    pub event_type: String,
    pub aggregate_id: String,
    pub available_at: DateTime<Utc>,
    pub locked_at: Option<DateTime<Utc>>,
    pub locked_by: Option<String>,
    pub processed_at: Option<DateTime<Utc>>,
    pub attempts: i32,
    pub last_error: Option<String>,
    pub created_at: DateTime<Utc>,
}

#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {}

impl ActiveModelBehavior for ActiveModel {}
