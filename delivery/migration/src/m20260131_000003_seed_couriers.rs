//! Seed Couriers Migration
//!
//! Seeds test couriers for MVP - available globally 24/7.

use sea_orm_migration::prelude::*;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        // Seed 10 MVP couriers - available everywhere, 24/7, all week
        manager
            .get_connection()
            .execute_unprepared(
                r#"
                INSERT INTO delivery.couriers (
                    name, phone, email, transport_type, max_distance_km,
                    work_zone, work_hours_start, work_hours_end, work_days
                ) VALUES
                -- Car couriers
                ('Max Müller', '+49 151 1234 0001', 'max.mueller@courier.de', 'car', 9999.00,
                 '*', '00:00', '23:59', ARRAY[1,2,3,4,5,6,7]),
                ('Hans Weber', '+49 151 1234 0003', 'hans.weber@courier.de', 'car', 9999.00,
                 '*', '00:00', '23:59', ARRAY[1,2,3,4,5,6,7]),
                ('Klaus Fischer', '+49 151 1234 0005', 'klaus.fischer@courier.de', 'car', 9999.00,
                 '*', '00:00', '23:59', ARRAY[1,2,3,4,5,6,7]),
                
                -- Motorcycle couriers
                ('Anna Schmidt', '+49 151 1234 0002', 'anna.schmidt@courier.de', 'motorcycle', 9999.00,
                 '*', '00:00', '23:59', ARRAY[1,2,3,4,5,6,7]),
                ('Maria Bauer', '+49 151 1234 0004', 'maria.bauer@courier.de', 'motorcycle', 9999.00,
                 '*', '00:00', '23:59', ARRAY[1,2,3,4,5,6,7]),
                ('Sophie Wagner', '+49 151 1234 0006', 'sophie.wagner@courier.de', 'motorcycle', 9999.00,
                 '*', '00:00', '23:59', ARRAY[1,2,3,4,5,6,7]),
                
                -- Bicycle couriers
                ('Thomas Braun', '+49 151 1234 0007', 'thomas.braun@courier.de', 'bicycle', 50.00,
                 '*', '00:00', '23:59', ARRAY[1,2,3,4,5,6,7]),
                ('Julia Hoffmann', '+49 151 1234 0008', 'julia.hoffmann@courier.de', 'bicycle', 50.00,
                 '*', '00:00', '23:59', ARRAY[1,2,3,4,5,6,7]),
                
                -- Walking couriers
                ('Markus Schulz', '+49 151 1234 0009', 'markus.schulz@courier.de', 'walking', 10.00,
                 '*', '00:00', '23:59', ARRAY[1,2,3,4,5,6,7]),
                ('Emma Klein', '+49 151 1234 0010', 'emma.klein@courier.de', 'walking', 10.00,
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
                    'hans.weber@courier.de',
                    'klaus.fischer@courier.de',
                    'anna.schmidt@courier.de',
                    'maria.bauer@courier.de',
                    'sophie.wagner@courier.de',
                    'thomas.braun@courier.de',
                    'julia.hoffmann@courier.de',
                    'markus.schulz@courier.de',
                    'emma.klein@courier.de'
                );
                "#,
            )
            .await?;

        Ok(())
    }
}
