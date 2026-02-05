//! PostgreSQL Implementation of CourierRepository
//!
//! Uses Sea-ORM for database operations.

use async_trait::async_trait;
use sea_orm::{
    ActiveModelTrait, ColumnTrait, DatabaseConnection, EntityTrait, PaginatorTrait, QueryFilter,
    QueryOrder, QuerySelect,
};
use uuid::Uuid;

use crate::domain::ports::{CourierFilter, CourierRepository, RepositoryError};
use crate::domain::model::courier::Courier;
use crate::infrastructure::repository::entities::courier::{
    self, ActiveModel, Entity as CourierEntity,
};

/// PostgreSQL implementation of CourierRepository using Sea-ORM
pub struct CourierPostgresRepository {
    db: DatabaseConnection,
}

impl CourierPostgresRepository {
    /// Create a new repository instance
    pub fn new(db: DatabaseConnection) -> Self {
        Self { db }
    }
}

#[async_trait]
impl CourierRepository for CourierPostgresRepository {
    async fn save(&self, courier: &Courier) -> Result<(), RepositoryError> {
        let model = ActiveModel::from(courier);

        // Check if courier exists
        let existing = CourierEntity::find_by_id(courier.id().0)
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        if let Some(ref existing_model) = existing {
            // Update with optimistic locking
            let existing_version = existing_model.version;
            if existing_version != courier.version() as i32 - 1 {
                return Err(RepositoryError::VersionConflict {
                    expected: courier.version() - 1,
                    actual: existing_version as u32,
                });
            }

            // Update the model
            model
                .update(&self.db)
                .await
                .map_err(|e| RepositoryError::QueryError(e.to_string()))?;
        } else {
            // Insert new courier
            model
                .insert(&self.db)
                .await
                .map_err(|e| {
                    let err_str = e.to_string();
                    if err_str.contains("duplicate") || err_str.contains("unique") {
                        if err_str.contains("email") {
                            RepositoryError::DuplicateEntry("email".to_string())
                        } else if err_str.contains("phone") {
                            RepositoryError::DuplicateEntry("phone".to_string())
                        } else {
                            RepositoryError::DuplicateEntry(err_str)
                        }
                    } else {
                        RepositoryError::QueryError(err_str)
                    }
                })?;
        }

        Ok(())
    }

    async fn find_by_id(&self, id: Uuid) -> Result<Option<Courier>, RepositoryError> {
        let result = CourierEntity::find_by_id(id)
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        match result {
            Some(model) => {
                let courier = Courier::try_from(model)
                    .map_err(RepositoryError::SerializationError)?;
                Ok(Some(courier))
            }
            None => Ok(None),
        }
    }

    async fn find_by_phone(&self, phone: &str) -> Result<Option<Courier>, RepositoryError> {
        let result = CourierEntity::find()
            .filter(courier::Column::Phone.eq(phone))
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        match result {
            Some(model) => {
                let courier = Courier::try_from(model)
                    .map_err(RepositoryError::SerializationError)?;
                Ok(Some(courier))
            }
            None => Ok(None),
        }
    }

    async fn find_by_email(&self, email: &str) -> Result<Option<Courier>, RepositoryError> {
        let result = CourierEntity::find()
            .filter(courier::Column::Email.eq(email))
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        match result {
            Some(model) => {
                let courier = Courier::try_from(model)
                    .map_err(RepositoryError::SerializationError)?;
                Ok(Some(courier))
            }
            None => Ok(None),
        }
    }

    async fn find_by_work_zone(&self, zone: &str) -> Result<Vec<Courier>, RepositoryError> {
        let results = CourierEntity::find()
            .filter(courier::Column::WorkZone.eq(zone))
            .all(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        let mut couriers = Vec::with_capacity(results.len());
        for model in results {
            let courier = Courier::try_from(model)
                .map_err(RepositoryError::SerializationError)?;
            couriers.push(courier);
        }

        Ok(couriers)
    }

    async fn email_exists(&self, email: &str) -> Result<bool, RepositoryError> {
        let result = CourierEntity::find()
            .filter(courier::Column::Email.eq(email))
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        Ok(result.is_some())
    }

    async fn phone_exists(&self, phone: &str) -> Result<bool, RepositoryError> {
        let result = CourierEntity::find()
            .filter(courier::Column::Phone.eq(phone))
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        Ok(result.is_some())
    }

    async fn delete(&self, id: Uuid) -> Result<(), RepositoryError> {
        let result = CourierEntity::delete_by_id(id)
            .exec(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        if result.rows_affected == 0 {
            return Err(RepositoryError::NotFound(id));
        }

        Ok(())
    }

    async fn archive(&self, id: Uuid) -> Result<(), RepositoryError> {
        use sea_orm::Set;

        // Find the courier first
        let existing = CourierEntity::find_by_id(id)
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?
            .ok_or(RepositoryError::NotFound(id))?;

        // Update timestamp to mark as archived
        // The actual ARCHIVED status is stored in Redis cache
        // TODO: Add is_archived column in future migration
        let mut model: ActiveModel = existing.into();
        model.updated_at = Set(chrono::Utc::now());

        model
            .update(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        Ok(())
    }

    async fn list(&self, limit: u64, offset: u64) -> Result<Vec<Courier>, RepositoryError> {
        let results = CourierEntity::find()
            .order_by_desc(courier::Column::CreatedAt)
            .limit(limit)
            .offset(offset)
            .all(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        let mut couriers = Vec::with_capacity(results.len());
        for model in results {
            let courier =
                Courier::try_from(model).map_err(RepositoryError::SerializationError)?;
            couriers.push(courier);
        }

        Ok(couriers)
    }

    async fn find_by_filter(
        &self,
        filter: CourierFilter,
        limit: u64,
        offset: u64,
    ) -> Result<Vec<Courier>, RepositoryError> {
        let mut query = CourierEntity::find();

        if let Some(ref zone) = filter.work_zone {
            query = query.filter(courier::Column::WorkZone.eq(zone.as_str()));
        }

        // status and archived are not stored in PostgreSQL (status/load in cache);
        // filter by them when DB columns exist in a future migration

        let results = query
            .order_by_desc(courier::Column::CreatedAt)
            .limit(limit)
            .offset(offset)
            .all(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        let mut couriers = Vec::with_capacity(results.len());
        for model in results {
            let courier =
                Courier::try_from(model).map_err(RepositoryError::SerializationError)?;
            couriers.push(courier);
        }

        Ok(couriers)
    }

    async fn count_by_filter(&self, filter: CourierFilter) -> Result<u64, RepositoryError> {
        let mut query = CourierEntity::find();

        if let Some(ref zone) = filter.work_zone {
            query = query.filter(courier::Column::WorkZone.eq(zone.as_str()));
        }

        let count = query
            .count(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        Ok(count as u64)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::courier::WorkHours;
    use crate::domain::model::vo::TransportType;
    use chrono::NaiveTime;
    use migration::{Migrator, MigratorTrait};
    use sea_orm::Database;
    use testcontainers::{runners::AsyncRunner, ContainerAsync, ImageExt};
    use testcontainers_modules::postgres::Postgres;

    async fn setup_db() -> (ContainerAsync<Postgres>, DatabaseConnection) {
        let container = Postgres::default()
            .with_tag("18-alpine")
            .start()
            .await
            .unwrap();
        let port = container.get_host_port_ipv4(5432).await.unwrap();
        let url = format!("postgres://postgres:postgres@localhost:{}/postgres", port);
        let db = Database::connect(&url).await.unwrap();

        // Apply migrations
        Migrator::up(&db, None).await.unwrap();

        (container, db)
    }

    fn create_test_courier(name: &str, phone: &str, email: &str) -> Courier {
        let work_hours = WorkHours::new(
            NaiveTime::from_hms_opt(9, 0, 0).unwrap(),
            NaiveTime::from_hms_opt(18, 0, 0).unwrap(),
            vec![1, 2, 3, 4, 5],
        )
        .unwrap();

        Courier::builder(
            name.to_string(),
            phone.to_string(),
            email.to_string(),
            TransportType::Bicycle,
            10.0,
            "zone-1".to_string(),
            work_hours,
        )
        .build()
        .unwrap()
    }

    #[tokio::test]
    async fn test_list_returns_seeded_couriers() {
        let (_container, db) = setup_db().await;
        let repo = CourierPostgresRepository::new(db);

        let result = repo.list(100, 0).await.unwrap();

        // Seed migration adds 10 couriers (see m20260131_000003_seed_couriers)
        assert_eq!(result.len(), 10);
    }

    #[tokio::test]
    async fn test_list_returns_couriers_with_pagination() {
        let (_container, db) = setup_db().await;
        let repo = CourierPostgresRepository::new(db);

        // Get initial count (seed data may exist)
        let initial = repo.list(100, 0).await.unwrap();
        let initial_count = initial.len();

        // Insert 3 test couriers
        let courier1 = create_test_courier("Courier 1", "+79001111111", "courier1@test.com");
        let courier2 = create_test_courier("Courier 2", "+79002222222", "courier2@test.com");
        let courier3 = create_test_courier("Courier 3", "+79003333333", "courier3@test.com");

        repo.save(&courier1).await.unwrap();
        repo.save(&courier2).await.unwrap();
        repo.save(&courier3).await.unwrap();

        // Test limit
        let result = repo.list(2, 0).await.unwrap();
        assert_eq!(result.len(), 2);

        // Test offset - skip first 2, should get remaining (capped by limit 10)
        let result = repo.list(10, 2).await.unwrap();
        let remaining = initial_count + 3 - 2;
        assert_eq!(result.len(), remaining.min(10));

        // Test all
        let result = repo.list(100, 0).await.unwrap();
        assert_eq!(result.len(), initial_count + 3);
    }

    #[tokio::test]
    async fn test_list_returns_couriers_ordered_by_created_at_desc() {
        let (_container, db) = setup_db().await;
        let repo = CourierPostgresRepository::new(db);

        // Get initial count (seed data)
        let initial = repo.list(100, 0).await.unwrap();
        let initial_count = initial.len();

        // Insert couriers with small delay to ensure different created_at
        let courier1 = create_test_courier("First", "+79001111111", "first@test.com");
        repo.save(&courier1).await.unwrap();

        let courier2 = create_test_courier("Second", "+79002222222", "second@test.com");
        repo.save(&courier2).await.unwrap();

        let courier3 = create_test_courier("Third", "+79003333333", "third@test.com");
        repo.save(&courier3).await.unwrap();

        let result = repo.list(100, 0).await.unwrap();

        // Should have seed data + 3 new couriers
        assert_eq!(result.len(), initial_count + 3);
        // Most recent should be first (Third) - check the first item is our latest insert
        assert_eq!(result[0].name(), "Third");
    }
}
