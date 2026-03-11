//! Add Recipient Contact Columns to Packages
//!
//! Adds recipient_name, recipient_phone, recipient_email to delivery.packages
//! for storing recipient contact details from OMS AcceptOrder.
//! Uses IF NOT EXISTS so the migration is idempotent when columns already exist.

use sea_orm_migration::prelude::*;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .get_connection()
            .execute_unprepared(
                "ALTER TABLE delivery.packages
                 ADD COLUMN IF NOT EXISTS recipient_name TEXT,
                 ADD COLUMN IF NOT EXISTS recipient_phone TEXT,
                 ADD COLUMN IF NOT EXISTS recipient_email TEXT",
            )
            .await?;

        Ok(())
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .get_connection()
            .execute_unprepared(
                "ALTER TABLE delivery.packages
                 DROP COLUMN IF EXISTS recipient_name,
                 DROP COLUMN IF EXISTS recipient_phone,
                 DROP COLUMN IF EXISTS recipient_email",
            )
            .await?;

        Ok(())
    }
}
