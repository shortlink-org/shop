//! Create Courier Location History Table Migration
//!
//! Creates the `geolocation.courier_location_history` table.

use sea_orm_migration::prelude::*;

use super::m20260201_000001_create_courier_current_locations::Geolocation;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        // Create courier_location_history table
        manager
            .create_table(
                Table::create()
                    .table((Geolocation::Schema, CourierLocationHistory::Table))
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
                    .col(ColumnDef::new(CourierLocationHistory::Speed).double())
                    .col(ColumnDef::new(CourierLocationHistory::Heading).double())
                    .col(
                        ColumnDef::new(CourierLocationHistory::RecordedAt)
                            .timestamp_with_time_zone()
                            .not_null()
                            .extra("DEFAULT NOW()"),
                    )
                    .to_owned(),
            )
            .await?;

        // Create index for courier_id + timestamp queries
        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_courier_location_history_courier_timestamp")
                    .table((Geolocation::Schema, CourierLocationHistory::Table))
                    .col(CourierLocationHistory::CourierId)
                    .col(CourierLocationHistory::Timestamp)
                    .to_owned(),
            )
            .await?;

        // Create index for recorded_at (for cleanup queries)
        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_courier_location_history_recorded_at")
                    .table((Geolocation::Schema, CourierLocationHistory::Table))
                    .col(CourierLocationHistory::RecordedAt)
                    .to_owned(),
            )
            .await?;

        Ok(())
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .drop_table(
                Table::drop()
                    .table((Geolocation::Schema, CourierLocationHistory::Table))
                    .to_owned(),
            )
            .await?;

        Ok(())
    }
}

/// Courier location history table columns
#[derive(Iden)]
pub enum CourierLocationHistory {
    Table,
    Id,
    CourierId,
    Latitude,
    Longitude,
    Accuracy,
    Timestamp,
    Speed,
    Heading,
    RecordedAt,
}
