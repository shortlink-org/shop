//! Delivery Service Database Migrations
//!
//! Sea-ORM migrations for the delivery schema.

pub use sea_orm_migration::prelude::*;

mod m20260131_000001_create_couriers;
mod m20260131_000002_create_packages;
mod m20260131_000003_seed_couriers;
mod m20260131_000004_create_courier_location_history;

pub struct Migrator;

#[async_trait::async_trait]
impl MigratorTrait for Migrator {
    fn migrations() -> Vec<Box<dyn MigrationTrait>> {
        vec![
            Box::new(m20260131_000001_create_couriers::Migration),
            Box::new(m20260131_000002_create_packages::Migration),
            Box::new(m20260131_000003_seed_couriers::Migration),
            Box::new(m20260131_000004_create_courier_location_history::Migration),
        ]
    }
}
