//! Create Couriers Table Migration
//!
//! Creates the `delivery.couriers` table for storing courier profiles.

use sea_orm_migration::prelude::*;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        // Create schema
        manager
            .get_connection()
            .execute_unprepared("CREATE SCHEMA IF NOT EXISTS delivery")
            .await?;

        // Create transport_type enum
        manager
            .get_connection()
            .execute_unprepared(
                "DO $$ BEGIN
                    CREATE TYPE delivery.transport_type AS ENUM ('walking', 'bicycle', 'motorcycle', 'car');
                EXCEPTION
                    WHEN duplicate_object THEN null;
                END $$;",
            )
            .await?;

        // Create couriers table
        manager
            .create_table(
                Table::create()
                    .table((Delivery::Schema, Couriers::Table))
                    .if_not_exists()
                    .col(
                        ColumnDef::new(Couriers::Id)
                            .uuid()
                            .not_null()
                            .primary_key()
                            .extra("DEFAULT gen_random_uuid()"),
                    )
                    .col(
                        ColumnDef::new(Couriers::Name)
                            .string_len(255)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Couriers::Phone)
                            .string_len(20)
                            .not_null()
                            .unique_key(),
                    )
                    .col(
                        ColumnDef::new(Couriers::Email)
                            .string_len(255)
                            .not_null()
                            .unique_key(),
                    )
                    .col(
                        ColumnDef::new(Couriers::TransportType)
                            .string_len(20)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Couriers::MaxDistanceKm)
                            .decimal_len(8, 2)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Couriers::WorkZone)
                            .string_len(100)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Couriers::WorkHoursStart)
                            .time()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Couriers::WorkHoursEnd)
                            .time()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Couriers::WorkDays)
                            .array(ColumnType::Integer)
                            .not_null(),
                    )
                    .col(ColumnDef::new(Couriers::PushToken).text())
                    .col(
                        ColumnDef::new(Couriers::CreatedAt)
                            .timestamp_with_time_zone()
                            .not_null()
                            .extra("DEFAULT NOW()"),
                    )
                    .col(
                        ColumnDef::new(Couriers::UpdatedAt)
                            .timestamp_with_time_zone()
                            .not_null()
                            .extra("DEFAULT NOW()"),
                    )
                    .col(
                        ColumnDef::new(Couriers::Version)
                            .integer()
                            .not_null()
                            .default(1),
                    )
                    .to_owned(),
            )
            .await?;

        // Create indexes
        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_couriers_work_zone")
                    .table((Delivery::Schema, Couriers::Table))
                    .col(Couriers::WorkZone)
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_couriers_transport_type")
                    .table((Delivery::Schema, Couriers::Table))
                    .col(Couriers::TransportType)
                    .to_owned(),
            )
            .await?;

        Ok(())
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        // Drop table
        manager
            .drop_table(
                Table::drop()
                    .table((Delivery::Schema, Couriers::Table))
                    .to_owned(),
            )
            .await?;

        // Drop enum type
        manager
            .get_connection()
            .execute_unprepared("DROP TYPE IF EXISTS delivery.transport_type")
            .await?;

        // Note: We don't drop the schema as other tables might use it

        Ok(())
    }
}

/// Schema identifier
#[derive(Iden)]
pub enum Delivery {
    #[iden = "delivery"]
    Schema,
}

/// Couriers table columns
#[derive(Iden)]
pub enum Couriers {
    Table,
    Id,
    Name,
    Phone,
    Email,
    TransportType,
    MaxDistanceKm,
    WorkZone,
    WorkHoursStart,
    WorkHoursEnd,
    WorkDays,
    PushToken,
    CreatedAt,
    UpdatedAt,
    Version,
}
