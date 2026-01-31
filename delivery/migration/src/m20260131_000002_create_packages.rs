//! Create Packages Table Migration
//!
//! Creates the `delivery.packages` table for storing package data.

use sea_orm_migration::prelude::*;

use super::m20260131_000001_create_couriers::Delivery;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        // Create package_status enum
        manager
            .get_connection()
            .execute_unprepared(
                "DO $$ BEGIN
                    CREATE TYPE delivery.package_status AS ENUM (
                        'accepted', 'in_pool', 'assigned', 'in_transit',
                        'delivered', 'not_delivered', 'requires_handling'
                    );
                EXCEPTION
                    WHEN duplicate_object THEN null;
                END $$;",
            )
            .await?;

        // Create priority enum
        manager
            .get_connection()
            .execute_unprepared(
                "DO $$ BEGIN
                    CREATE TYPE delivery.priority AS ENUM ('normal', 'urgent');
                EXCEPTION
                    WHEN duplicate_object THEN null;
                END $$;",
            )
            .await?;

        // Create packages table
        manager
            .create_table(
                Table::create()
                    .table((Delivery::Schema, Packages::Table))
                    .if_not_exists()
                    .col(
                        ColumnDef::new(Packages::Id)
                            .uuid()
                            .not_null()
                            .primary_key()
                            .extra("DEFAULT gen_random_uuid()"),
                    )
                    .col(ColumnDef::new(Packages::OrderId).uuid().not_null())
                    .col(ColumnDef::new(Packages::CustomerId).uuid().not_null())
                    // Pickup address
                    .col(
                        ColumnDef::new(Packages::PickupStreet)
                            .string_len(255)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Packages::PickupCity)
                            .string_len(100)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Packages::PickupPostalCode)
                            .string_len(20)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Packages::PickupLatitude)
                            .double()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Packages::PickupLongitude)
                            .double()
                            .not_null(),
                    )
                    // Delivery address
                    .col(
                        ColumnDef::new(Packages::DeliveryStreet)
                            .string_len(255)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Packages::DeliveryCity)
                            .string_len(100)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Packages::DeliveryPostalCode)
                            .string_len(20)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Packages::DeliveryLatitude)
                            .double()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Packages::DeliveryLongitude)
                            .double()
                            .not_null(),
                    )
                    // Delivery period
                    .col(
                        ColumnDef::new(Packages::DeliveryPeriodStart)
                            .timestamp_with_time_zone()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Packages::DeliveryPeriodEnd)
                            .timestamp_with_time_zone()
                            .not_null(),
                    )
                    // Package info
                    .col(
                        ColumnDef::new(Packages::WeightKg)
                            .decimal_len(8, 3)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Packages::Priority)
                            .string_len(20)
                            .not_null()
                            .default("normal"),
                    )
                    .col(
                        ColumnDef::new(Packages::Status)
                            .string_len(20)
                            .not_null()
                            .default("accepted"),
                    )
                    .col(ColumnDef::new(Packages::CourierId).uuid())
                    .col(
                        ColumnDef::new(Packages::Zone)
                            .string_len(100)
                            .not_null(),
                    )
                    // Timestamps
                    .col(
                        ColumnDef::new(Packages::CreatedAt)
                            .timestamp_with_time_zone()
                            .not_null()
                            .extra("DEFAULT NOW()"),
                    )
                    .col(
                        ColumnDef::new(Packages::UpdatedAt)
                            .timestamp_with_time_zone()
                            .not_null()
                            .extra("DEFAULT NOW()"),
                    )
                    .col(ColumnDef::new(Packages::AssignedAt).timestamp_with_time_zone())
                    .col(ColumnDef::new(Packages::DeliveredAt).timestamp_with_time_zone())
                    .col(ColumnDef::new(Packages::NotDeliveredReason).text())
                    .col(
                        ColumnDef::new(Packages::Version)
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
                    .name("idx_packages_order_id")
                    .table((Delivery::Schema, Packages::Table))
                    .col(Packages::OrderId)
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_packages_status")
                    .table((Delivery::Schema, Packages::Table))
                    .col(Packages::Status)
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_packages_courier_id")
                    .table((Delivery::Schema, Packages::Table))
                    .col(Packages::CourierId)
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_packages_zone")
                    .table((Delivery::Schema, Packages::Table))
                    .col(Packages::Zone)
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_packages_priority_created")
                    .table((Delivery::Schema, Packages::Table))
                    .col(Packages::Priority)
                    .col(Packages::CreatedAt)
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
                    .table((Delivery::Schema, Packages::Table))
                    .to_owned(),
            )
            .await?;

        // Drop enum types
        manager
            .get_connection()
            .execute_unprepared("DROP TYPE IF EXISTS delivery.package_status")
            .await?;

        manager
            .get_connection()
            .execute_unprepared("DROP TYPE IF EXISTS delivery.priority")
            .await?;

        Ok(())
    }
}

/// Packages table columns
#[derive(Iden)]
pub enum Packages {
    Table,
    Id,
    OrderId,
    CustomerId,
    PickupStreet,
    PickupCity,
    PickupPostalCode,
    PickupLatitude,
    PickupLongitude,
    DeliveryStreet,
    DeliveryCity,
    DeliveryPostalCode,
    DeliveryLatitude,
    DeliveryLongitude,
    DeliveryPeriodStart,
    DeliveryPeriodEnd,
    WeightKg,
    Priority,
    Status,
    CourierId,
    Zone,
    CreatedAt,
    UpdatedAt,
    AssignedAt,
    DeliveredAt,
    NotDeliveredReason,
    Version,
}
