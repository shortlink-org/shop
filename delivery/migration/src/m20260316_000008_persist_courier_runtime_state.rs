//! Persist Courier Runtime State
//!
//! Moves courier runtime attributes from Redis-only storage into PostgreSQL so
//! Redis can act as a cache instead of the source of truth.

use sea_orm_migration::prelude::*;

use crate::m20260131_000001_create_couriers::{Couriers, Delivery};

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .alter_table(
                Table::alter()
                    .table((Delivery::Schema, Couriers::Table))
                    .add_column(
                        ColumnDef::new(Couriers::Status)
                            .string_len(20)
                            .not_null()
                            .default("unavailable"),
                    )
                    .add_column(
                        ColumnDef::new(Couriers::CurrentLoad)
                            .integer()
                            .not_null()
                            .default(0),
                    )
                    .add_column(
                        ColumnDef::new(Couriers::Rating)
                            .double()
                            .not_null()
                            .default(0.0),
                    )
                    .add_column(
                        ColumnDef::new(Couriers::SuccessfulDeliveries)
                            .integer()
                            .not_null()
                            .default(0),
                    )
                    .add_column(
                        ColumnDef::new(Couriers::FailedDeliveries)
                            .integer()
                            .not_null()
                            .default(0),
                    )
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_couriers_status")
                    .table((Delivery::Schema, Couriers::Table))
                    .col(Couriers::Status)
                    .to_owned(),
            )
            .await?;

        Ok(())
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .drop_index(
                Index::drop()
                    .name("idx_couriers_status")
                    .table((Delivery::Schema, Couriers::Table))
                    .to_owned(),
            )
            .await?;

        manager
            .alter_table(
                Table::alter()
                    .table((Delivery::Schema, Couriers::Table))
                    .drop_column(Couriers::FailedDeliveries)
                    .drop_column(Couriers::SuccessfulDeliveries)
                    .drop_column(Couriers::Rating)
                    .drop_column(Couriers::CurrentLoad)
                    .drop_column(Couriers::Status)
                    .to_owned(),
            )
            .await?;

        Ok(())
    }
}
