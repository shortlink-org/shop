//! Add Recipient Contact Columns to Packages
//!
//! Adds recipient_name, recipient_phone, recipient_email to delivery.packages
//! for storing recipient contact details from OMS AcceptOrder.

use sea_orm_migration::prelude::*;

use super::m20260131_000001_create_couriers::Delivery;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .alter_table(
                Table::alter()
                    .table((Delivery::Schema, Packages::Table))
                    .add_column(
                        ColumnDef::new(Packages::RecipientName)
                            .text()
                            .null(),
                    )
                    .add_column(
                        ColumnDef::new(Packages::RecipientPhone)
                            .text()
                            .null(),
                    )
                    .add_column(
                        ColumnDef::new(Packages::RecipientEmail)
                            .text()
                            .null(),
                    )
                    .to_owned(),
            )
            .await?;

        Ok(())
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .alter_table(
                Table::alter()
                    .table((Delivery::Schema, Packages::Table))
                    .drop_column(Packages::RecipientName)
                    .drop_column(Packages::RecipientPhone)
                    .drop_column(Packages::RecipientEmail)
                    .to_owned(),
            )
            .await?;

        Ok(())
    }
}

/// Packages table columns (subset used in this migration)
#[derive(Iden)]
pub enum Packages {
    Table,
    RecipientName,
    RecipientPhone,
    RecipientEmail,
}
