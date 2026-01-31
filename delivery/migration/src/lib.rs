//! Delivery Service Database Migrations
//!
//! Sea-ORM migrations for the delivery schema.

pub use sea_orm_migration::prelude::*;

mod m20260131_000001_create_couriers;

pub struct Migrator;

#[async_trait::async_trait]
impl MigratorTrait for Migrator {
    fn migrations() -> Vec<Box<dyn MigrationTrait>> {
        vec![Box::new(m20260131_000001_create_couriers::Migration)]
    }
}
