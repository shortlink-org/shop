//! PostgreSQL Implementation of CourierRepository
//!
//! Uses Sea-ORM for database operations.

use async_trait::async_trait;
use sea_orm::{ActiveModelTrait, ColumnTrait, DatabaseConnection, EntityTrait, QueryFilter};
use uuid::Uuid;

use crate::boundary::ports::{CourierRepository, RepositoryError};
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

        if existing.is_some() {
            // Update with optimistic locking
            let existing_version = existing.as_ref().unwrap().version;
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
                    .map_err(|e| RepositoryError::SerializationError(e))?;
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
                    .map_err(|e| RepositoryError::SerializationError(e))?;
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
                    .map_err(|e| RepositoryError::SerializationError(e))?;
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
                .map_err(|e| RepositoryError::SerializationError(e))?;
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
}
