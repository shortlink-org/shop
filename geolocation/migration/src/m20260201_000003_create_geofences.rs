//! Create Geofences Table Migration
//!
//! Creates the `geolocation.geofences` table.

use sea_orm_migration::prelude::*;

use super::m20260201_000001_create_courier_current_locations::Geolocation;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        // Create geofence_shape enum
        manager
            .get_connection()
            .execute_unprepared(
                "DO $$ BEGIN
                    CREATE TYPE geolocation.geofence_shape AS ENUM ('circle', 'polygon', 'rectangle');
                EXCEPTION
                    WHEN duplicate_object THEN null;
                END $$;",
            )
            .await?;

        // Create geofence_trigger enum
        manager
            .get_connection()
            .execute_unprepared(
                "DO $$ BEGIN
                    CREATE TYPE geolocation.geofence_trigger AS ENUM ('on_enter', 'on_exit', 'on_both');
                EXCEPTION
                    WHEN duplicate_object THEN null;
                END $$;",
            )
            .await?;

        // Create geofences table
        manager
            .create_table(
                Table::create()
                    .table((Geolocation::Schema, Geofences::Table))
                    .if_not_exists()
                    .col(
                        ColumnDef::new(Geofences::Id)
                            .uuid()
                            .not_null()
                            .primary_key()
                            .extra("DEFAULT gen_random_uuid()"),
                    )
                    .col(
                        ColumnDef::new(Geofences::Name)
                            .string_len(255)
                            .not_null(),
                    )
                    .col(ColumnDef::new(Geofences::Description).text())
                    .col(
                        ColumnDef::new(Geofences::ShapeType)
                            .string_len(20)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Geofences::ShapeData)
                            .json_binary()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Geofences::TriggerType)
                            .string_len(20)
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(Geofences::IsActive)
                            .boolean()
                            .not_null()
                            .default(true),
                    )
                    .col(
                        ColumnDef::new(Geofences::CreatedAt)
                            .timestamp_with_time_zone()
                            .not_null()
                            .extra("DEFAULT NOW()"),
                    )
                    .col(
                        ColumnDef::new(Geofences::UpdatedAt)
                            .timestamp_with_time_zone()
                            .not_null()
                            .extra("DEFAULT NOW()"),
                    )
                    .to_owned(),
            )
            .await?;

        // Create index for active geofences
        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_geofences_is_active")
                    .table((Geolocation::Schema, Geofences::Table))
                    .col(Geofences::IsActive)
                    .to_owned(),
            )
            .await?;

        Ok(())
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .drop_table(
                Table::drop()
                    .table((Geolocation::Schema, Geofences::Table))
                    .to_owned(),
            )
            .await?;

        // Drop enum types
        manager
            .get_connection()
            .execute_unprepared("DROP TYPE IF EXISTS geolocation.geofence_trigger")
            .await?;

        manager
            .get_connection()
            .execute_unprepared("DROP TYPE IF EXISTS geolocation.geofence_shape")
            .await?;

        Ok(())
    }
}

/// Geofences table columns
#[derive(Iden)]
pub enum Geofences {
    Table,
    Id,
    Name,
    Description,
    ShapeType,
    ShapeData,
    TriggerType,
    IsActive,
    CreatedAt,
    UpdatedAt,
}
