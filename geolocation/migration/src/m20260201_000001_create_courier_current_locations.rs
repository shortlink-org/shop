//! Create Courier Current Locations Table Migration
//!
//! Creates the `geolocation.courier_current_locations` table.

use sea_orm_migration::prelude::*;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        // Create schema
        manager
            .get_connection()
            .execute_unprepared("CREATE SCHEMA IF NOT EXISTS geolocation")
            .await?;

        // Create courier_current_locations table
        manager
            .create_table(
                Table::create()
                    .table((Geolocation::Schema, CourierCurrentLocations::Table))
                    .if_not_exists()
                    .col(
                        ColumnDef::new(CourierCurrentLocations::CourierId)
                            .uuid()
                            .not_null()
                            .primary_key(),
                    )
                    .col(
                        ColumnDef::new(CourierCurrentLocations::Latitude)
                            .double()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(CourierCurrentLocations::Longitude)
                            .double()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(CourierCurrentLocations::Accuracy)
                            .double()
                            .not_null(),
                    )
                    .col(
                        ColumnDef::new(CourierCurrentLocations::Timestamp)
                            .timestamp_with_time_zone()
                            .not_null(),
                    )
                    .col(ColumnDef::new(CourierCurrentLocations::Speed).double())
                    .col(ColumnDef::new(CourierCurrentLocations::Heading).double())
                    .col(
                        ColumnDef::new(CourierCurrentLocations::UpdatedAt)
                            .timestamp_with_time_zone()
                            .not_null()
                            .extra("DEFAULT NOW()"),
                    )
                    .to_owned(),
            )
            .await?;

        // Create spatial index (using btree on lat/lon for now, PostGIS would use GIST)
        manager
            .create_index(
                Index::create()
                    .if_not_exists()
                    .name("idx_courier_current_locations_coords")
                    .table((Geolocation::Schema, CourierCurrentLocations::Table))
                    .col(CourierCurrentLocations::Latitude)
                    .col(CourierCurrentLocations::Longitude)
                    .to_owned(),
            )
            .await?;

        Ok(())
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .drop_table(
                Table::drop()
                    .table((Geolocation::Schema, CourierCurrentLocations::Table))
                    .to_owned(),
            )
            .await?;

        Ok(())
    }
}

/// Schema identifier
#[derive(Iden)]
pub enum Geolocation {
    #[iden = "geolocation"]
    Schema,
}

/// Courier current locations table columns
#[derive(Iden)]
pub enum CourierCurrentLocations {
    Table,
    CourierId,
    Latitude,
    Longitude,
    Accuracy,
    Timestamp,
    Speed,
    Heading,
    UpdatedAt,
}
