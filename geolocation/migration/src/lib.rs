//! Geolocation Service Database Migrations
//!
//! Sea-ORM migrations for the geolocation schema.

pub use sea_orm_migration::prelude::*;

mod m20260201_000001_create_courier_current_locations;
mod m20260201_000002_create_courier_location_history;
mod m20260201_000003_create_geofences;

pub struct Migrator;

#[async_trait::async_trait]
impl MigratorTrait for Migrator {
    fn migrations() -> Vec<Box<dyn MigrationTrait>> {
        vec![
            Box::new(m20260201_000001_create_courier_current_locations::Migration),
            Box::new(m20260201_000002_create_courier_location_history::Migration),
            Box::new(m20260201_000003_create_geofences::Migration),
        ]
    }
}
