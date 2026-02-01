//! Create Courier Location History Table Migration
//!
//! Creates the `delivery.courier_location_history` table for storing
//! courier GPS location history for tracking and analytics.

use sea_orm_migration::prelude::*;

use super::m20260131_000001_create_couriers::{Couriers, Delivery};

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        // Create courier_location_history table
        manager
            .create_table(
                Table::create()
                    .table((Delivery::Schema, CourierLocationHistory::Table))
                    .if_not_exists()
                    .col(
                        ColumnDef::new(CourierLocationHistory::Id)
                            .uuid()
                            .not_null()
                            .primary_key()
                            .extra("DEFAULT gen_random_uuid()"),
                    )
                    .col(
                        ColumnDef::new(CourierLocationHistory::CourierId)
                            .uuid()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(CourierLocationHistory::Latitude)
                            .double()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(CourierLocationHistory::Longitude)
                            .double()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(CourierLocationHistory::Accuracy)
                            .double()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(CourierLocationHistory::Timestamp)
                            .timestamp_with_time_zone()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(CourierLocationHistory::Speed)
                            .double()
                            .null(),
                    )
                    .col(
                        ColumnDef::new(CourierLocationHistory::Heading)
                            .double()
                            .null(),
                    )
                    .col(
                        ColumnDef::new(CourierLocationHistory::CreatedAt)
                            .timestamp_with_time_zone()
                            .not_null()
                            .extra("DEFAULT NOW()"),
                    )
                    .foreign_key(
                        ForeignKey::create()
                            .name("fk_location_history_courier")
                            .from(
                                (Delivery::Schema, CourierLocationHistory::Table),
                                CourierLocationHistory::CourierId,
                            )
                            .to((Delivery::Schema, Couriers::Table), Couriers::Id)
                            .on_delete(ForeignKeyAction::Cascade),
                    )
                    .to_owned(),
            )
            .await?;

        // Create index on courier_id + timestamp for efficient history queries
        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_location_history_courier_timestamp")
                    .table((Delivery::Schema, CourierLocationHistory::Table))
                    .col(CourierLocationHistory::CourierId)
                    .col(CourierLocationHistory::Timestamp)
                    .to_owned(),
            )
            .await?;

        // Create index on timestamp for cleanup queries
        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_location_history_timestamp")
                    .table((Delivery::Schema, CourierLocationHistory::Table))
                    .col(CourierLocationHistory::Timestamp)
                    .to_owned(),
            )
            .await?;

        // Create index on created_at for retention policy
        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_location_history_created_at")
                    .table((Delivery::Schema, CourierLocationHistory::Table))
                    .col(CourierLocationHistory::CreatedAt)
                    .to_owned(),
            )
            .await?;

        Ok(())
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .drop_table(
                Table::drop()
                    .table((Delivery::Schema, CourierLocationHistory::Table))
                    .to_owned(),
            )
            .await?;

        Ok(())
    }
}

/// Courier Location History table columns
#[derive(Iden)]
pub enum CourierLocationHistory {
    #[iden = "courier_location_history"]
    Table,
    Id,
    CourierId,
    Latitude,
    Longitude,
    Accuracy,
    Timestamp,
    Speed,
    Heading,
    CreatedAt,
}
