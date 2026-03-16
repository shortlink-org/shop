//! Replace legacy string not_delivered_reason with structured columns.
//!
//! The service is still new, so no backfill or compatibility path is kept.

use sea_orm_migration::prelude::*;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .get_connection()
            .execute_unprepared(
                r#"
                ALTER TABLE delivery.packages
                    ADD COLUMN IF NOT EXISTS not_delivered_reason_code VARCHAR(50),
                    ADD COLUMN IF NOT EXISTS not_delivered_reason_description TEXT;

                ALTER TABLE delivery.packages
                    DROP COLUMN IF EXISTS not_delivered_reason;

                CREATE INDEX IF NOT EXISTS idx_packages_not_delivered_reason_code
                    ON delivery.packages (not_delivered_reason_code);
                "#,
            )
            .await?;

        Ok(())
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .get_connection()
            .execute_unprepared(
                r#"
                DROP INDEX IF EXISTS delivery.idx_packages_not_delivered_reason_code;

                ALTER TABLE delivery.packages
                    DROP COLUMN IF EXISTS not_delivered_reason_description,
                    DROP COLUMN IF EXISTS not_delivered_reason_code,
                    ADD COLUMN IF NOT EXISTS not_delivered_reason TEXT;
                "#,
            )
            .await?;

        Ok(())
    }
}
