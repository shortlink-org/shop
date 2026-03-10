//! Create Outbox Messages Table Migration
//!
//! Creates the `delivery.outbox_messages` table used for transactional outbox.

use sea_orm_migration::prelude::*;

use super::m20260131_000001_create_couriers::Delivery;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .create_table(
                Table::create()
                    .table((Delivery::Schema, OutboxMessages::Table))
                    .if_not_exists()
                    .col(
                        ColumnDef::new(OutboxMessages::Id)
                            .uuid()
                            .not_null()
                            .primary_key()
                            .extra("DEFAULT gen_random_uuid()"),
                    )
                    .col(ColumnDef::new(OutboxMessages::Topic).text().not_null())
                    .col(ColumnDef::new(OutboxMessages::MessageKey).text().not_null())
                    .col(ColumnDef::new(OutboxMessages::Payload).binary().not_null())
                    .col(
                        ColumnDef::new(OutboxMessages::Headers)
                            .text()
                            .not_null()
                            .default("{}"),
                    )
                    .col(
                        ColumnDef::new(OutboxMessages::PayloadEncoding)
                            .string_len(32)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(OutboxMessages::EventType)
                            .string_len(128)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(OutboxMessages::AggregateId)
                            .text()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(OutboxMessages::AvailableAt)
                            .timestamp_with_time_zone()
                            .not_null()
                            .extra("DEFAULT NOW()"),
                    )
                    .col(ColumnDef::new(OutboxMessages::LockedAt).timestamp_with_time_zone())
                    .col(ColumnDef::new(OutboxMessages::LockedBy).text())
                    .col(ColumnDef::new(OutboxMessages::ProcessedAt).timestamp_with_time_zone())
                    .col(
                        ColumnDef::new(OutboxMessages::Attempts)
                            .integer()
                            .not_null()
                            .default(0),
                    )
                    .col(ColumnDef::new(OutboxMessages::LastError).text())
                    .col(
                        ColumnDef::new(OutboxMessages::CreatedAt)
                            .timestamp_with_time_zone()
                            .not_null()
                            .extra("DEFAULT NOW()"),
                    )
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_outbox_messages_pending")
                    .table((Delivery::Schema, OutboxMessages::Table))
                    .col(OutboxMessages::ProcessedAt)
                    .col(OutboxMessages::AvailableAt)
                    .col(OutboxMessages::LockedAt)
                    .col(OutboxMessages::CreatedAt)
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_outbox_messages_created_at")
                    .table((Delivery::Schema, OutboxMessages::Table))
                    .col(OutboxMessages::CreatedAt)
                    .to_owned(),
            )
            .await?;

        Ok(())
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .drop_table(
                Table::drop()
                    .table((Delivery::Schema, OutboxMessages::Table))
                    .to_owned(),
            )
            .await?;

        Ok(())
    }
}

#[derive(Iden)]
enum OutboxMessages {
    #[iden = "outbox_messages"]
    Table,
    Id,
    Topic,
    #[iden = "message_key"]
    MessageKey,
    Payload,
    Headers,
    #[iden = "payload_encoding"]
    PayloadEncoding,
    #[iden = "event_type"]
    EventType,
    #[iden = "aggregate_id"]
    AggregateId,
    #[iden = "available_at"]
    AvailableAt,
    #[iden = "locked_at"]
    LockedAt,
    #[iden = "locked_by"]
    LockedBy,
    #[iden = "processed_at"]
    ProcessedAt,
    Attempts,
    #[iden = "last_error"]
    LastError,
    #[iden = "created_at"]
    CreatedAt,
}
