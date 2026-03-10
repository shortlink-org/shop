//! Add Customer Phone to Packages
//!
//! Adds the missing customer_phone column expected by the Package entity.

use sea_orm_migration::prelude::*;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .get_connection()
            .execute_unprepared(
                "ALTER TABLE delivery.packages ADD COLUMN IF NOT EXISTS customer_phone TEXT",
            )
            .await?;

        Ok(())
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .get_connection()
            .execute_unprepared(
                "ALTER TABLE delivery.packages DROP COLUMN IF EXISTS customer_phone",
            )
            .await?;

        Ok(())
    }
}
