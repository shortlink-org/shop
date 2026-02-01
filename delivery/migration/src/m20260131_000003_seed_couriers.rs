//! Seed Couriers Migration
//!
//! Seeds test couriers for MVP - available globally 24/7.

use sea_orm_migration::prelude::*;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        // Seed 2 MVP couriers - available everywhere, 24/7, all week
        manager
            .get_connection()
            .execute_unprepared(
                r#"
                INSERT INTO delivery.couriers (
                    name, phone, email, transport_type, max_distance_km,
                    work_zone, work_hours_start, work_hours_end, work_days
                ) VALUES
                -- MVP courier 1 - car, unlimited range
                ('Max Müller', '+49 151 1234 0001', 'max.mueller@courier.de', 'car', 9999.00,
                 '*', '00:00', '23:59', ARRAY[1,2,3,4,5,6,7]),
                
                -- MVP courier 2 - motorcycle, unlimited range
                ('Anna Schmidt', '+49 151 1234 0002', 'anna.schmidt@courier.de', 'motorcycle', 9999.00,
                 '*', '00:00', '23:59', ARRAY[1,2,3,4,5,6,7])
                
                ON CONFLICT DO NOTHING;
                "#,
            )
            .await?;

        Ok(())
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        // Remove seeded couriers by their unique emails
        manager
            .get_connection()
            .execute_unprepared(
                r#"
                DELETE FROM delivery.couriers
                WHERE email IN (
                    'max.mueller@courier.de',
                    'anna.schmidt@courier.de'
                );
                "#,
            )
            .await?;

        Ok(())
    }
}
